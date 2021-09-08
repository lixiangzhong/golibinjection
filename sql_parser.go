package golibinjection

import (
	"fmt"
	"strings"
)

type ptr_lookup_fn func(state *sqli_state, lookuptype int, word []byte) int
type pt2Function func(sf *sfilter) int

type sqli_state struct {
	/*
	 * input, does not need to be null terminated.
	 * it is also not modified.
	 */
	s string

	/*
	 * input length
	 */
	slen int

	/*
	 * How to lookup a word or fingerprint
	 */
	lookup   ptr_lookup_fn
	userdata interface{}

	/*
	 *
	 */
	flags int

	/*
	 * pos is the index in the string during tokenization
	 */
	pos int

	tokenvec [8]sqli_token

	/*
	 * Pointer to token position in tokenvec, above
	 */
	current *sqli_token

	/*
	 * fingerprint pattern c-string
	 * +1 for ending null
	 * Minimum of 8 bytes to add gcc's -fstack-protector to work
	 */
	fingerprint []byte

	/*
	 * Line number of code that said decided if the input was SQLi or
	 * not.  Most of the time it's line that said "it's not a matching
	 * fingerprint" but there is other logic that sometimes approves
	 * an input. This is only useful for debugging.
	 *
	 */
	reason int

	/* Number of ddw (dash-dash-white) comments
	 * These comments are in the form of
	 *   '--[whitespace]' or '--[EOF]'
	 *
	 * All databases treat this as a comment.
	 */
	stats_comment_ddw int

	/* Number of ddx (dash-dash-[notwhite]) comments
	 *
	 * ANSI SQL treats these are comments, MySQL treats this as
	 * two unary operators '-' '-'
	 *
	 * If you are parsing result returns FALSE and
	 * stats_comment_dd > 0, you should reparse with
	 * COMMENT_MYSQL
	 *
	 */
	stats_comment_ddx int

	/*
	 * c-style comments found  /x .. x/
	 */
	stats_comment_c int

	/* '#' operators or MySQL EOL comments found
	 *
	 */
	stats_comment_hash int

	/*
	 * number of tokens folded away
	 */
	stats_folds int

	/*
	 * total tokens processed
	 */
	stats_tokens int
}

const (
	LIBINJECTION_SQLI_MAX_TOKENS = 5
)

type sfilter = sqli_state

func SQLInject(src string) error {
	state := sqli_state{}
	libinjection_sqli_init(&state, src, len(src), 0)
	issqli := libinjection_is_sqli(&state)
	if issqli {
		return fmt.Errorf("内容非法，涵盖非法信息，指纹:%s", string(state.fingerprint))
	}

	return nil
}

func libinjection_sqli_tokenize(sf *sqli_state) bool {
	current := sf.current
	s := sf.s
	slen := sf.slen

	if slen == 0 {
		return false
	}

	st_clear(current)
	sf.current = current

	/*
	 * if we are at beginning of string
	 *  and in single-quote or double quote mode
	 *  then pretend the input starts with a quote
	 */
	if sf.pos == 0 && (sf.flags&(FLAG_QUOTE_SINGLE|FLAG_QUOTE_DOUBLE)) > 0 {
		sf.pos = parse_string_core(s, slen, 0, current, flag2delim(sf.flags), 0)
		sf.stats_tokens += 1
		return true
	}

	for sf.pos < slen {

		/*
		 * get current character
		 */
		ch := s[sf.pos]

		/*
		 * look up the parser, and call it
		 *
		 * Porting Note: this is mapping of char to function
		 *   charparsers[ch]()
		 */
		fnptr := tokenCharMap[ch]

		sf.pos = fnptr(sf)

		/*
		 *
		 */
		if current.ttype != CHAR_NULL {
			sf.stats_tokens += 1
			return true
		}
	}
	return false
}

func libinjection_sqli_init(sf *sqli_state, s string, len, flags int) {
	if flags == 0 {
		flags = FLAG_QUOTE_NONE | FLAG_SQL_ANSI
	}

	sf.pos = 0
	sf.stats_tokens = 0
	sf.stats_folds = 0

	sf.fingerprint = []byte{}
	sf.s = s
	sf.slen = len
	sf.lookup = libinjection_sqli_lookup_word
	sf.userdata = 0
	sf.flags = flags
	sf.current = &sf.tokenvec[0]
}

func libinjection_sqli_reset(sf *sqli_state, flags int) {
	userdata := sf.userdata
	lookup := sf.lookup

	if flags == 0 {
		flags = FLAG_QUOTE_NONE | FLAG_SQL_ANSI
	}
	libinjection_sqli_init(sf, sf.s, sf.slen, flags)
	sf.lookup = lookup
	sf.userdata = userdata
}

func libinjection_sqli_callback(sf *sqli_state, fn ptr_lookup_fn, userdata interface{}) {
	if fn == nil {
		sf.lookup = libinjection_sqli_lookup_word
		sf.userdata = nil
	} else {
		sf.lookup = fn
		sf.userdata = userdata
	}
}

/** See if two tokens can be merged since they are compound SQL phrases.
 *
 * This takes two tokens, and, if they are the right type,
 * merges their values together.  Then checks to see if the
 * new value is special using the PHRASES mapping.
 *
 * Example: "UNION" + "ALL" ==> "UNION ALL"
 *
 * C Security Notes: this is safe to use C-strings (null-terminated)
 *  since the types involved by definition do not have embedded nulls
 *  (e.g. there is no keyword with embedded null)
 *
 * Porting Notes: since this is C, it's oddly complicated.
 *  This is just:  multikeywords[token.value + ' ' + token2.value]
 *
 */
func syntax_merge_words(sf *sqli_state, a, b *sqli_token) bool {
	var sz1 int
	var sz2 int
	var sz3 int
	tmp := &strings.Builder{}
	var ch uint8

	/* first token is of right type? */
	if !(a.ttype == TYPE_KEYWORD ||
		a.ttype == TYPE_BAREWORD ||
		a.ttype == TYPE_OPERATOR ||
		a.ttype == TYPE_UNION ||
		a.ttype == TYPE_FUNCTION ||
		a.ttype == TYPE_EXPRESSION ||
		a.ttype == TYPE_TSQL ||
		a.ttype == TYPE_SQLTYPE) {
		return false
	}

	if !(b.ttype == TYPE_KEYWORD ||
		b.ttype == TYPE_BAREWORD ||
		b.ttype == TYPE_OPERATOR ||
		b.ttype == TYPE_UNION ||
		b.ttype == TYPE_FUNCTION ||
		b.ttype == TYPE_EXPRESSION ||
		b.ttype == TYPE_TSQL ||
		b.ttype == TYPE_SQLTYPE ||
		b.ttype == TYPE_LOGIC_OPERATOR) {
		return false
	}

	sz1 = a.len
	sz2 = b.len
	sz3 = sz1 + sz2 + 1                      /* +1 for space in the middle */
	if sz3 >= LIBINJECTION_SQLI_TOKEN_SIZE { /* make sure there is room for ending null */
		return false
	}
	/*
	 * oddly annoying  last.val + ' ' + current.val
	 */
	tmp.Write([]byte(a.val[:sz1]))
	tmp.Write([]byte{' '})
	tmp.Write([]byte(b.val[:sz2]))
	ch = uint8(sf.lookup(sf, LOOKUP_WORD, []byte(tmp.String())))
	if ch != CHAR_NULL {
		st_assign(a, ch, a.pos, sz3, tmp.String())
		return true
	} else {
		return false
	}
}

func libinjection_sqli_fold(sf *sqli_state) int {
	var last_comment sqli_token

	/* POS is the position of where the NEXT token goes */
	pos := 0

	/* LEFT is a count of how many tokens that are already
	folded or processed (i.e. part of the fingerprint) */
	left := 0

	more := true

	st_clear(&last_comment)

	/* Skip all initial comments, right-parens ( and unary operators
	 *
	 */
	sf.current = &sf.tokenvec[0]
	for more {
		more = libinjection_sqli_tokenize(sf)
		if !(sf.current.ttype == TYPE_COMMENT ||
			sf.current.ttype == TYPE_LEFTPARENS ||
			sf.current.ttype == TYPE_SQLTYPE ||
			st_is_unary_op(sf.current)) {
			break
		}
	}

	if !more {
		/* If input was only comments, unary or (, then exit */
		return 0
	} else {
		/* it's some other token */
		pos += 1
	}

	for {
		/* do we have all the max number of tokens?  if so do
		 * some special cases for 5 tokens
		 */
		if pos >= LIBINJECTION_SQLI_MAX_TOKENS {
			if (sf.tokenvec[0].ttype == TYPE_NUMBER &&
				(sf.tokenvec[1].ttype == TYPE_OPERATOR || sf.tokenvec[1].ttype == TYPE_COMMA) &&
				sf.tokenvec[2].ttype == TYPE_LEFTPARENS &&
				sf.tokenvec[3].ttype == TYPE_NUMBER &&
				sf.tokenvec[4].ttype == TYPE_RIGHTPARENS) ||
				(sf.tokenvec[0].ttype == TYPE_BAREWORD &&
					sf.tokenvec[1].ttype == TYPE_OPERATOR &&
					sf.tokenvec[2].ttype == TYPE_LEFTPARENS &&
					(sf.tokenvec[3].ttype == TYPE_BAREWORD || sf.tokenvec[3].ttype == TYPE_NUMBER) &&
					sf.tokenvec[4].ttype == TYPE_RIGHTPARENS) ||
				(sf.tokenvec[0].ttype == TYPE_NUMBER &&
					sf.tokenvec[1].ttype == TYPE_RIGHTPARENS &&
					sf.tokenvec[2].ttype == TYPE_COMMA &&
					sf.tokenvec[3].ttype == TYPE_LEFTPARENS &&
					sf.tokenvec[4].ttype == TYPE_NUMBER) ||
				(sf.tokenvec[0].ttype == TYPE_BAREWORD &&
					sf.tokenvec[1].ttype == TYPE_RIGHTPARENS &&
					sf.tokenvec[2].ttype == TYPE_OPERATOR &&
					sf.tokenvec[3].ttype == TYPE_LEFTPARENS &&
					sf.tokenvec[4].ttype == TYPE_BAREWORD) {
				if pos > LIBINJECTION_SQLI_MAX_TOKENS {
					st_copy(&sf.tokenvec[1], &sf.tokenvec[LIBINJECTION_SQLI_MAX_TOKENS])
					pos = 2
					left = 0
				} else {
					pos = 1
					left = 0
				}
			}
		}

		if !more || left >= LIBINJECTION_SQLI_MAX_TOKENS {
			left = pos
			break
		}

		/* get up to two tokens */
		for more && pos <= LIBINJECTION_SQLI_MAX_TOKENS && (pos-left) < 2 {
			sf.current = &sf.tokenvec[pos]
			more = libinjection_sqli_tokenize(sf)
			if more {
				if sf.current.ttype == TYPE_COMMENT {
					st_copy(&last_comment, sf.current)
				} else {
					last_comment.ttype = CHAR_NULL
					pos += 1
				}
			}
		}

		/* did we get 2 tokens? if not then we are done */
		if pos-left < 2 {
			left = pos
			continue
		}

		/* FOLD: "ss" . "s"
		 * "foo" "bar" is valid SQL
		 * just ignore second string
		 */
		if sf.tokenvec[left].ttype == TYPE_STRING && sf.tokenvec[left+1].ttype == TYPE_STRING {
			pos -= 1
			sf.stats_folds += 1
			continue
		} else if sf.tokenvec[left].ttype == TYPE_SEMICOLON && sf.tokenvec[left+1].ttype == TYPE_SEMICOLON {
			/* not sure how various engines handle
			 * 'select 1;;drop table foo' or
			 * 'select 1; /x foo x/; drop table foo'
			 * to prevent surprises, just fold away repeated semicolons
			 */
			pos -= 1
			sf.stats_folds += 1
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_OPERATOR ||
			sf.tokenvec[left].ttype == TYPE_LOGIC_OPERATOR) &&
			(st_is_unary_op(&sf.tokenvec[left+1]) ||
				sf.tokenvec[left+1].ttype == TYPE_SQLTYPE) {
			pos -= 1
			sf.stats_folds += 1
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_LEFTPARENS &&
			st_is_unary_op(&sf.tokenvec[left+1]) {
			pos -= 1
			sf.stats_folds += 1
			if left > 0 {
				left -= 1
			}
			continue
		} else if syntax_merge_words(sf, &sf.tokenvec[left], &sf.tokenvec[left+1]) {
			pos -= 1
			sf.stats_folds += 1
			if left > 0 {
				left -= 1
			}
			continue
		} else if sf.tokenvec[left].ttype == TYPE_SEMICOLON &&
			sf.tokenvec[left+1].ttype == TYPE_FUNCTION &&
			(sf.tokenvec[left+1].val[0] == 'I' ||
				sf.tokenvec[left+1].val[0] == 'i') &&
			(sf.tokenvec[left+1].val[1] == 'F' ||
				sf.tokenvec[left+1].val[1] == 'f') {
			/* IF is normally a function, except in Transact-SQL where it can be used as a
			 * standalone control flow operator, e.g. ; IF 1=1 ...
			 * if found after a semicolon, convert from 'f' type to 'T' type
			 */
			sf.tokenvec[left+1].ttype = TYPE_TSQL
			/* left += 2; */
			continue /* reparse everything, but we probably can advance left, and pos */
		} else if (sf.tokenvec[left].ttype == TYPE_BAREWORD || sf.tokenvec[left].ttype == TYPE_VARIABLE) &&
			sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS && (
		/* TSQL functions but common enough to be column names */
		cstrcasecmp("USER_ID", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("USER_NAME", sf.tokenvec[left].val) == 0 ||

			/* Function in MYSQL */
			cstrcasecmp("DATABASE", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("PASSWORD", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("USER", sf.tokenvec[left].val) == 0 ||

			/* Mysql words that act as a variable and are a function */

			/* TSQL current_users is fake-variable */
			/* http://msdn.microsoft.com/en-us/library/ms176050.aspx */
			cstrcasecmp("CURRENT_USER", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("CURRENT_DATE", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("CURRENT_TIME", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("CURRENT_TIMESTAMP", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("LOCALTIME", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("LOCALTIMESTAMP", sf.tokenvec[left].val) == 0) {

			/* pos is the same
			 * other conversions need to go here... for instance
			 * password CAN be a function, coalesce CAN be a function
			 */
			sf.tokenvec[left].ttype = TYPE_FUNCTION
			continue
		} else if sf.tokenvec[left].ttype == TYPE_KEYWORD && (cstrcasecmp("IN", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("NOT IN", sf.tokenvec[left].val) == 0) {

			if sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS {
				/* got .... IN ( ...  (or 'NOT IN')
				 * it's an operator
				 */
				sf.tokenvec[left].ttype = TYPE_OPERATOR
			} else {
				/*
				 * it's a nothing
				 */
				sf.tokenvec[left].ttype = TYPE_BAREWORD
			}

			/* "IN" can be used as "IN BOOLEAN MODE" for mysql
			 *  in which case merging of words can be done later
			 * other wise it acts as an equality operator __ IN (values..)
			 *
			 * here we got "IN" "(" so it's an operator.
			 * also back track to handle "NOT IN"
			 * might need to do the same with like
			 * two use cases   "foo" LIKE "BAR" (normal operator)
			 *  "foo" = LIKE(1,2)
			 */
			continue
		} else if sf.tokenvec[left].ttype == TYPE_OPERATOR && (cstrcasecmp("LIKE", sf.tokenvec[left].val) == 0 ||
			cstrcasecmp("NOT LIKE", sf.tokenvec[left].val) == 0) {

			if sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS {
				/* SELECT LIKE(...
				 * it's a function
				 */
				sf.tokenvec[left].ttype = TYPE_FUNCTION
			}
		} else if sf.tokenvec[left].ttype == TYPE_SQLTYPE &&
			(sf.tokenvec[left+1].ttype == TYPE_BAREWORD ||
				sf.tokenvec[left+1].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+1].ttype == TYPE_SQLTYPE ||
				sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS ||
				sf.tokenvec[left+1].ttype == TYPE_FUNCTION ||
				sf.tokenvec[left+1].ttype == TYPE_VARIABLE ||
				sf.tokenvec[left+1].ttype == TYPE_STRING) {
			st_copy(&sf.tokenvec[left], &sf.tokenvec[left+1])
			pos -= 1
			sf.stats_folds += 1
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_COLLATE &&
			sf.tokenvec[left+1].ttype == TYPE_BAREWORD {
			/*
			 * there are too many collation types.. so if the bareword has a "_"
			 * then it's TYPE_SQLTYPE
			 */
			if strchr(sf.tokenvec[left+1].val, '_') != 0 {
				sf.tokenvec[left+1].ttype = TYPE_SQLTYPE
				left = 0
			}
		} else if sf.tokenvec[left].ttype == TYPE_BACKSLASH {
			if st_is_arithmetic_op(&sf.tokenvec[left+1]) {
				/* very weird case in TSQL where '\%1' is parsed as '0 % 1', etc */
				sf.tokenvec[left].ttype = TYPE_NUMBER
			} else {
				/* just ignore it.. Again T-SQL seems to parse \1 as "1" */
				st_copy(&sf.tokenvec[left], &sf.tokenvec[left+1])
				pos -= 1
				sf.stats_folds += 1
			}
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_LEFTPARENS &&
			sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS {
			pos -= 1
			left = 0
			sf.stats_folds += 1
			continue
		} else if sf.tokenvec[left].ttype == TYPE_RIGHTPARENS &&
			sf.tokenvec[left+1].ttype == TYPE_RIGHTPARENS {
			pos -= 1
			left = 0
			sf.stats_folds += 1
			continue
		} else if sf.tokenvec[left].ttype == TYPE_LEFTBRACE &&
			sf.tokenvec[left+1].ttype == TYPE_BAREWORD {

			/*
			 * MySQL Degenerate case --
			 *
			 *   select { ``.``.id };  -- valid !!!
			 *   select { ``.``.``.id };  -- invalid
			 *   select ``.``.id; -- invalid
			 *   select { ``.id }; -- invalid
			 *
			 * so it appears {``.``.id} is a magic case
			 * I suspect this is "current database, current table, field id"
			 *
			 * The folding code can't look at more than 3 tokens, and
			 * I don't want to make two passes.
			 *
			 * Since "{ ``" so rare, we are just going to blacklist it.
			 *
			 * Highly likely this will need revisiting!
			 *
			 * CREDIT @rsalgado 2013-11-25
			 */
			if sf.tokenvec[left+1].len == 0 {
				sf.tokenvec[left+1].ttype = TYPE_EVIL
				return (left + 2)
			}
			/* weird ODBC / MYSQL  {foo expr} -. expr
			 * but for this rule we just strip away the "{ foo" part
			 */
			left = 0
			pos -= 2
			sf.stats_folds += 2
			continue
		} else if sf.tokenvec[left+1].ttype == TYPE_RIGHTBRACE {
			pos -= 1
			left = 0
			sf.stats_folds += 1
			continue
		}

		/* all cases of handing 2 tokens is done
		and nothing matched.  Get one more token
		*/
		for more && pos <= LIBINJECTION_SQLI_MAX_TOKENS && pos-left < 3 {
			sf.current = &(sf.tokenvec[pos])
			more = libinjection_sqli_tokenize(sf)
			if more {
				if sf.current.ttype == TYPE_COMMENT {
					st_copy(&last_comment, sf.current)
				} else {
					last_comment.ttype = CHAR_NULL
					pos += 1
				}
			}
		}

		/* do we have three tokens? If not then we are done */
		if pos-left < 3 {
			left = pos
			continue
		}

		/*
		 * now look for three token folding
		 */
		if sf.tokenvec[left].ttype == TYPE_NUMBER &&
			sf.tokenvec[left+1].ttype == TYPE_OPERATOR &&
			sf.tokenvec[left+2].ttype == TYPE_NUMBER {
			pos -= 2
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_OPERATOR &&
			sf.tokenvec[left+1].ttype != TYPE_LEFTPARENS &&
			sf.tokenvec[left+2].ttype == TYPE_OPERATOR {
			left = 0
			pos -= 2
			continue
		} else if sf.tokenvec[left].ttype == TYPE_LOGIC_OPERATOR &&
			sf.tokenvec[left+2].ttype == TYPE_LOGIC_OPERATOR {
			pos -= 2
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_VARIABLE &&
			sf.tokenvec[left+1].ttype == TYPE_OPERATOR &&
			(sf.tokenvec[left+2].ttype == TYPE_VARIABLE ||
				sf.tokenvec[left+2].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+2].ttype == TYPE_BAREWORD) {
			pos -= 2
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_BAREWORD ||
			sf.tokenvec[left].ttype == TYPE_NUMBER) &&
			sf.tokenvec[left+1].ttype == TYPE_OPERATOR &&
			(sf.tokenvec[left+2].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+2].ttype == TYPE_BAREWORD) {
			pos -= 2
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_BAREWORD ||
			sf.tokenvec[left].ttype == TYPE_NUMBER ||
			sf.tokenvec[left].ttype == TYPE_VARIABLE ||
			sf.tokenvec[left].ttype == TYPE_STRING) &&
			sf.tokenvec[left+1].ttype == TYPE_OPERATOR &&
			streq(sf.tokenvec[left+1].val, "::") &&
			sf.tokenvec[left+2].ttype == TYPE_SQLTYPE {
			pos -= 2
			left = 0
			sf.stats_folds += 2
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_BAREWORD ||
			sf.tokenvec[left].ttype == TYPE_NUMBER ||
			sf.tokenvec[left].ttype == TYPE_STRING ||
			sf.tokenvec[left].ttype == TYPE_VARIABLE) &&
			sf.tokenvec[left+1].ttype == TYPE_COMMA &&
			(sf.tokenvec[left+2].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+2].ttype == TYPE_BAREWORD ||
				sf.tokenvec[left+2].ttype == TYPE_STRING ||
				sf.tokenvec[left+2].ttype == TYPE_VARIABLE) {
			pos -= 2
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_EXPRESSION ||
			sf.tokenvec[left].ttype == TYPE_GROUP ||
			sf.tokenvec[left].ttype == TYPE_COMMA) &&
			st_is_unary_op(&sf.tokenvec[left+1]) &&
			sf.tokenvec[left+2].ttype == TYPE_LEFTPARENS {
			/* got something like SELECT + (, LIMIT + (
			 * remove unary operator
			 */
			st_copy(&sf.tokenvec[left+1], &sf.tokenvec[left+2])
			pos -= 1
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_KEYWORD ||
			sf.tokenvec[left].ttype == TYPE_EXPRESSION ||
			sf.tokenvec[left].ttype == TYPE_GROUP) &&
			st_is_unary_op(&sf.tokenvec[left+1]) &&
			(sf.tokenvec[left+2].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+2].ttype == TYPE_BAREWORD ||
				sf.tokenvec[left+2].ttype == TYPE_VARIABLE ||
				sf.tokenvec[left+2].ttype == TYPE_STRING ||
				sf.tokenvec[left+2].ttype == TYPE_FUNCTION) {
			/* remove unary operators
			 * select - 1
			 */
			st_copy(&sf.tokenvec[left+1], &sf.tokenvec[left+2])
			pos -= 1
			left = 0
			continue
		} else if sf.tokenvec[left].ttype == TYPE_COMMA &&
			st_is_unary_op(&sf.tokenvec[left+1]) &&
			(sf.tokenvec[left+2].ttype == TYPE_NUMBER ||
				sf.tokenvec[left+2].ttype == TYPE_BAREWORD ||
				sf.tokenvec[left+2].ttype == TYPE_VARIABLE ||
				sf.tokenvec[left+2].ttype == TYPE_STRING) {
			/*
			 * interesting case    turn ", -1"  .> ",1" PLUS we need to back up
			 * one token if possible to see if more folding can be done
			 * "1,-1" -. "1"
			 */
			st_copy(&sf.tokenvec[left+1], &sf.tokenvec[left+2])
			left = 0
			/* pos is >= 3 so this is safe */
			pos -= 3
			continue
		} else if sf.tokenvec[left].ttype == TYPE_COMMA &&
			st_is_unary_op(&sf.tokenvec[left+1]) &&
			sf.tokenvec[left+2].ttype == TYPE_FUNCTION {

			/* Separate case from above since you end up with
			 * 1,-sin(1) -. 1 (1)
			 * Here, just do
			 * 1,-sin(1) -. 1,sin(1)
			 * just remove unary operator
			 */
			st_copy(&sf.tokenvec[left+1], &sf.tokenvec[left+2])
			pos -= 1
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_BAREWORD) &&
			(sf.tokenvec[left+1].ttype == TYPE_DOT) &&
			(sf.tokenvec[left+2].ttype == TYPE_BAREWORD) {
			/* ignore the '.n'
			 * typically is this databasename.table
			 */
			pos -= 2
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_EXPRESSION) &&
			(sf.tokenvec[left+1].ttype == TYPE_DOT) &&
			(sf.tokenvec[left+2].ttype == TYPE_BAREWORD) {
			/* select . `foo` -. select `foo` */
			st_copy(&sf.tokenvec[left+1], &sf.tokenvec[left+2])
			pos -= 1
			left = 0
			continue
		} else if (sf.tokenvec[left].ttype == TYPE_FUNCTION) &&
			(sf.tokenvec[left+1].ttype == TYPE_LEFTPARENS) &&
			(sf.tokenvec[left+2].ttype != TYPE_RIGHTPARENS) {
			/*
			 * whats going on here
			 * Some SQL functions like USER() have 0 args
			 * if we get User(foo), then User is not a function
			 * This should be expanded since it eliminated a lot of false
			 * positives.
			 */
			if cstrcasecmp("USER", sf.tokenvec[left].val) == 0 {
				sf.tokenvec[left].ttype = TYPE_BAREWORD
			}
		}

		/* no folding -- assume left-most token is
		is good, now use the existing 2 tokens --
		do not get another
		*/

		left += 1

	} /* while(1) */

	/* if we have 4 or less tokens, and we had a comment token
	 * at the end, add it back
	 */

	if left < LIBINJECTION_SQLI_MAX_TOKENS && last_comment.ttype == TYPE_COMMENT {
		st_copy(&sf.tokenvec[left], &last_comment)
		left += 1
	}

	/* sometimes we grab a 6th token to help
	determine the type of token 5.
	*/
	if left > LIBINJECTION_SQLI_MAX_TOKENS {
		left = LIBINJECTION_SQLI_MAX_TOKENS
	}

	return left
}

/* secondary api: detects SQLi in a string, GIVEN a context.
 *
 * A context can be:
 *   *  CHAR_NULL (\0), process as is
 *   *  CHAR_SINGLE ('), process pretending input started with a
 *          single quote.
 *   *  CHAR_DOUBLE ("), process pretending input started with a
 *          double quote.
 *
 */
func libinjection_sqli_fingerprint(sql_state *sqli_state, flags int) string {
	tlen := 0

	libinjection_sqli_reset(sql_state, flags)

	tlen = libinjection_sqli_fold(sql_state)

	/* Check for magic PHP backquote comment
	 * If:
	 * * last token is of type "bareword"
	 * * And is quoted in a backtick
	 * * And isn't closed
	 * * And it's empty?
	 * Then convert it to comment
	 */
	if tlen > 2 &&
		sql_state.tokenvec[tlen-1].ttype == TYPE_BAREWORD &&
		sql_state.tokenvec[tlen-1].str_open == CHAR_TICK &&
		sql_state.tokenvec[tlen-1].len == 0 &&
		sql_state.tokenvec[tlen-1].str_close == CHAR_NULL {
		sql_state.tokenvec[tlen-1].ttype = TYPE_COMMENT
	}

	for i := 0; i < tlen; i++ {
		sql_state.fingerprint = append(sql_state.fingerprint, sql_state.tokenvec[i].ttype)
	}

	/*
	 * make the fingerprint pattern a c-string (null delimited)
	 */
	//sql_state.fingerprint[tlen] = CHAR_NULL

	/*
	 * check for 'X' in pattern, and then
	 * clear out all tokens
	 *
	 * this means parsing could not be done
	 * accurately due to pgsql's double comments
	 * or other syntax that isn't consistent.
	 * Should be very rare false positive
	 */
	if strchr(string(sql_state.fingerprint), TYPE_EVIL) > 0 {
		/*  needed for SWIG */
		sql_state.tokenvec[0].val = ""
		sql_state.fingerprint = []byte{TYPE_EVIL}
		sql_state.tokenvec[0].ttype = TYPE_EVIL
		sql_state.tokenvec[0].val = string(TYPE_EVIL)
		sql_state.tokenvec[1].ttype = CHAR_NULL
	}

	return string(sql_state.fingerprint)
}

func libinjection_sqli_check_fingerprint(sql_state *sqli_state) bool {
	return libinjection_sqli_blacklist(sql_state) && libinjection_sqli_not_whitelist(sql_state)
}

func libinjection_sqli_lookup_word(state *sqli_state, lookuptype int, word []byte) int {
	if lookuptype == LOOKUP_FINGERPRINT {
		if libinjection_sqli_check_fingerprint(state) {
			return 'X'
		}
		return 0
	} else {
		return int(bsearch_keyword_type(string(word)))
	}
}

func libinjection_sqli_blacklist(sql_state *sqli_state) bool {
	/*
	 * use minimum of 8 bytes to make sure gcc -fstack-protector
	 * works correctly
	 */
	len := len(string(sql_state.fingerprint))

	if len < 1 {
		sql_state.reason = 0
		return false
	}

	/*
	   to keep everything compatible, convert the
	   v0 fingerprint pattern to v1
	   v0: up to 5 chars, mixed case
	   v1: 1 char is '0', up to 5 more chars, upper case
	*/
	fp2 := make([]byte, 1)
	fp2[0] = '0'
	for i := 0; i < len; i++ {
		ch := sql_state.fingerprint[i]
		if ch >= 'a' && ch <= 'z' {
			ch -= 0x20
		}
		fp2 = append(fp2, ch)
	}

	patmatch := bsearch_keyword_type(string(fp2)) == TYPE_FINGERPRINT

	/*
	 * No match.
	 *
	 * Set sql_state.reason to current line number
	 * only for debugging purposes.
	 */
	if !patmatch {
		sql_state.reason = 0
		return false
	}

	return true
}

/*
 * return TRUE if SQLi, false is benign
 */
func libinjection_sqli_not_whitelist(state *sqli_state) bool {
	/*
	 * We assume we got a SQLi match
	 * This next part just helps reduce false positives.
	 *
	 */
	tlen := len(string(state.fingerprint))

	if tlen > 1 && state.fingerprint[tlen-1] == TYPE_COMMENT {
		/*
		 * if ending comment is contains 'sp_password' then it's SQLi!
		 * MS Audit log apparently ignores anything with
		 * 'sp_password' in it. Unable to find primary reference to
		 * this "feature" of SQL Server but seems to be known SQLi
		 * technique
		 */
		if my_memmem(state.s, "sp_password") > 0 {
			state.reason = 0
			return true
		}
	}

	switch tlen {
	case 2:
		{
			/*
			 * case 2 are "very small SQLi" which make them
			 * hard to tell from normal input...
			 */

			if state.fingerprint[1] == TYPE_UNION {
				if state.stats_tokens == 2 {
					/* not sure why but 1U comes up in SQLi attack
					 * likely part of parameter splitting/etc.
					 * lots of reasons why "1 union" might be normal
					 * input, so beep only if other SQLi things are present
					 */
					/* it really is a number and 'union'
					 * other wise it has folding or comments
					 */
					state.reason = 0
					return false
				} else {
					state.reason = 0
					return true
				}
			}
			/*
			 * if 'comment' is '#' ignore.. too many FP
			 */
			if state.tokenvec[1].val[0] == '#' {
				state.reason = 0
				return false
			}

			/*
			 * for fingerprint like 'nc', only comments of /x are treated
			 * as SQL... ending comments of "--" and "#" are not SQLi
			 */
			if state.tokenvec[0].ttype == TYPE_BAREWORD &&
				state.tokenvec[1].ttype == TYPE_COMMENT &&
				state.tokenvec[1].val[0] != '/' {
				state.reason = 0
				return false
			}

			/*
			 * if '1c' ends with '/x' then it's SQLi
			 */
			if state.tokenvec[0].ttype == TYPE_NUMBER &&
				state.tokenvec[1].ttype == TYPE_COMMENT &&
				state.tokenvec[1].val[0] == '/' {
				return true
			}

			/**
			 * there are some odd base64-looking query string values
			 * 1234-ABCDEFEhfhihwuefi--
			 * which evaluate to "1c"... these are not SQLi
			 * but 1234-- probably is.
			 * Make sure the "1" in "1c" is actually a true decimal number
			 *
			 * Need to check -original- string since the folding step
			 * may have merged tokens, e.g. "1+FOO" is folded into "1"
			 *
			 * Note: evasion: 1*1--
			 */
			if state.tokenvec[0].ttype == TYPE_NUMBER &&
				state.tokenvec[1].ttype == TYPE_COMMENT {
				if state.stats_tokens > 2 {
					/* we have some folding going on, highly likely SQLi */
					state.reason = 0
					return true
				}
				/*
				 * we check that next character after the number is either whitespace,
				 * or '/' or a '-' ==> SQLi.
				 */
				ch := state.s[state.tokenvec[0].len]
				if ch <= 32 {
					/* next char was whitespace,e.g. "1234 --"
					 * this isn't exactly correct.. ideally we should skip over all whitespace
					 * but this seems to be ok for now
					 */
					return true
				}
				if ch == '/' && state.s[state.tokenvec[0].len+1] == '*' {
					return true
				}
				if ch == '-' && state.s[state.tokenvec[0].len+1] == '-' {
					return true
				}

				state.reason = 0
				return false
			}

			/*
			 * detect obvious SQLi scans.. many people put '--' in plain text
			 * so only detect if input ends with '--', e.g. 1-- but not 1-- foo
			 */
			if state.tokenvec[1].len > 2 && state.tokenvec[1].val[0] == '-' {
				state.reason = 0
				return false
			}

			break
		} /* case 2 */
	case 3:
		{
			/*
			 * ...foo' + 'bar...
			 * no opening quote, no closing quote
			 * and each string has data
			 */

			fp := string(state.fingerprint)

			if streq(fp, "sos") || streq(fp, "s&s") {

				if (state.tokenvec[0].str_open == CHAR_NULL) && (state.tokenvec[2].str_close == CHAR_NULL) && (state.tokenvec[0].str_close == state.tokenvec[2].str_open) {
					/*
					 * if ....foo" + "bar....
					 */
					state.reason = 0
					return true
				}
				if state.stats_tokens == 3 {
					state.reason = 0
					return false
				}

				/*
				 * not SQLi
				 */
				state.reason = 0
				return false
			} else if streq(fp, "s&n") ||
				streq(fp, "n&1") ||
				streq(fp, "1&1") ||
				streq(fp, "1&v") ||
				streq(fp, "1&s") {
				/* 'sexy and 17' not SQLi
				 * 'sexy and 17<18'  SQLi
				 */
				if state.stats_tokens == 3 {
					state.reason = 0
					return false
				}
			} else if state.tokenvec[1].ttype == TYPE_KEYWORD {
				if state.tokenvec[1].len < 5 || cstrcasecmp("INTO", state.tokenvec[1].val[:4]) > 0 {
					/* if it's not "INTO OUTFILE", or "INTO DUMPFILE" (MySQL)
					 * then treat as safe
					 */
					state.reason = 0
					return false
				}
			}
			break
		} /* case 3 */
	case 4:
	case 5:
		{
			/* nothing right now */
			break
		} /* case 5 */
	} /* end switch */

	return true
}

/**  Main API, detects SQLi in an input.
 *
 *
 */
func reparse_as_mysql(sql_state *sqli_state) bool {
	return sql_state.stats_comment_ddx > 0 || sql_state.stats_comment_hash > 0
}

/*
 * This function is mostly use with SWIG
 */
func libinjection_sqli_get_token(sql_state *sqli_state, i int) *sqli_token {
	if i < 0 || i > LIBINJECTION_SQLI_MAX_TOKENS {
		return nil
	}
	return &sql_state.tokenvec[i]
}

func libinjection_is_sqli(sql_state *sqli_state) bool {
	s := sql_state.s
	slen := sql_state.slen

	/*
	 * no input? not SQLi
	 */
	if slen == 0 {
		return false
	}

	/*
	 * test input "as-is"
	 */
	libinjection_sqli_fingerprint(sql_state, FLAG_QUOTE_NONE|FLAG_SQL_ANSI)
	if sql_state.lookup(sql_state, LOOKUP_FINGERPRINT, sql_state.fingerprint) > 0 {
		return true
	} else if reparse_as_mysql(sql_state) {
		libinjection_sqli_fingerprint(sql_state, FLAG_QUOTE_NONE|FLAG_SQL_MYSQL)
		if sql_state.lookup(sql_state, LOOKUP_FINGERPRINT, sql_state.fingerprint) > 0 {
			return true
		}
	}

	/*
	 * if input has a single_quote, then
	 * test as if input was actually '
	 * example: if input if "1' = 1", then pretend it's
	 *   "'1' = 1"
	 * Porting Notes: example the same as doing
	 *   is_string_sqli(sql_state, "'" + s, slen+1, NULL, fn, arg)
	 *
	 */
	if memchr(s, 0, CHAR_SINGLE) > 0 {
		libinjection_sqli_fingerprint(sql_state, FLAG_QUOTE_SINGLE|FLAG_SQL_ANSI)
		if sql_state.lookup(sql_state, LOOKUP_FINGERPRINT, sql_state.fingerprint) > 0 {
			return true
		} else if reparse_as_mysql(sql_state) {
			libinjection_sqli_fingerprint(sql_state, FLAG_QUOTE_SINGLE|FLAG_SQL_MYSQL)
			if sql_state.lookup(sql_state, LOOKUP_FINGERPRINT, sql_state.fingerprint) > 0 {
				return true
			}
		}
	}

	/*
	 * same as above but with a double-quote "
	 */
	if memchr(s, 0, CHAR_DOUBLE) > 0 {
		libinjection_sqli_fingerprint(sql_state, FLAG_QUOTE_DOUBLE|FLAG_SQL_MYSQL)
		if sql_state.lookup(sql_state, LOOKUP_FINGERPRINT, sql_state.fingerprint) > 0 {
			return true
		}
	}

	/*
	 * Hurray, input is not SQLi
	 */
	return false
}
