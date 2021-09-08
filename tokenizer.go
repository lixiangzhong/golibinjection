package golibinjection

const (
	TYPE_TK_NONE        = 0
	TYPE_KEYWORD        = 'k'
	TYPE_UNION          = 'U'
	TYPE_GROUP          = 'B'
	TYPE_EXPRESSION     = 'E'
	TYPE_SQLTYPE        = 't'
	TYPE_FUNCTION       = 'f'
	TYPE_BAREWORD       = 'n'
	TYPE_NUMBER         = '1'
	TYPE_VARIABLE       = 'v'
	TYPE_STRING         = 's'
	TYPE_OPERATOR       = 'o'
	TYPE_LOGIC_OPERATOR = '&'
	TYPE_COMMENT        = 'c'
	TYPE_COLLATE        = 'A'
	TYPE_LEFTPARENS     = '('
	TYPE_RIGHTPARENS    = ')' /* not used? */
	TYPE_LEFTBRACE      = '{'
	TYPE_RIGHTBRACE     = '}'
	TYPE_DOT            = '.'
	TYPE_COMMA          = ','
	TYPE_COLON          = ':'
	TYPE_SEMICOLON      = ';'
	TYPE_TSQL           = 'T' /* TSQL start */
	TYPE_UNKNOWN        = '?'
	TYPE_EVIL           = 'X' /* unparsable, abort  */
	TYPE_FINGERPRINT    = 'F' /* not really a token */
	TYPE_BACKSLASH      = '\\'
)

const (
	FLAG_NONE         = 0
	FLAG_QUOTE_NONE   = 1 /* 1 << 0 */
	FLAG_QUOTE_SINGLE = 2 /* 1 << 1 */
	FLAG_QUOTE_DOUBLE = 4 /* 1 << 2 */

	FLAG_SQL_ANSI  = 8  /* 1 << 3 */
	FLAG_SQL_MYSQL = 16 /* 1 << 4 */
)

const (
	LOOKUP_WORD        = 1
	LOOKUP_TYPE        = 2
	LOOKUP_OPERATOR    = 3
	LOOKUP_FINGERPRINT = 4
)

const (
	LIBINJECTION_SQLI_TOKEN_SIZE = 32
)

type sqli_token struct {
	/*
	 * position and length of token
	 * in original string
	 */
	pos int
	len int

	/*  count:
	 *  in type 'v', used for number of opening '@'
	 *  but maybe used in other contexts
	 */
	count int

	ttype     byte
	str_open  uint8
	str_close uint8
	val       string
}

func ISDIGIT(s uint8) bool {
	return (s - '0') <= 9
}

func flag2delim(flag int) uint8 {
	if (flag & FLAG_QUOTE_SINGLE) > 0 {
		return CHAR_SINGLE
	} else if (flag & FLAG_QUOTE_DOUBLE) > 0 {
		return CHAR_DOUBLE
	}

	return CHAR_NULL
}

type parserFunc func(sf *sqli_state) int

var (
	tokenCharMap = [256]parserFunc{
		parse_white,     /* 0 */
		parse_white,     /* 1 */
		parse_white,     /* 2 */
		parse_white,     /* 3 */
		parse_white,     /* 4 */
		parse_white,     /* 5 */
		parse_white,     /* 6 */
		parse_white,     /* 7 */
		parse_white,     /* 8 */
		parse_white,     /* 9 */
		parse_white,     /* 10 */
		parse_white,     /* 11 */
		parse_white,     /* 12 */
		parse_white,     /* 13 */
		parse_white,     /* 14 */
		parse_white,     /* 15 */
		parse_white,     /* 16 */
		parse_white,     /* 17 */
		parse_white,     /* 18 */
		parse_white,     /* 19 */
		parse_white,     /* 20 */
		parse_white,     /* 21 */
		parse_white,     /* 22 */
		parse_white,     /* 23 */
		parse_white,     /* 24 */
		parse_white,     /* 25 */
		parse_white,     /* 26 */
		parse_white,     /* 27 */
		parse_white,     /* 28 */
		parse_white,     /* 29 */
		parse_white,     /* 30 */
		parse_white,     /* 31 */
		parse_white,     /* 32 */
		parse_operator2, /* 33 */
		parse_string,    /* 34 */
		parse_hash,      /* 35 */
		parse_money,     /* 36 */
		parse_operator1, /* 37 */
		parse_operator2, /* 38 */
		parse_string,    /* 39 */
		parse_char,      /* 40 */
		parse_char,      /* 41 */
		parse_operator2, /* 42 */
		parse_operator1, /* 43 */
		parse_char,      /* 44 */
		parse_dash,      /* 45 */
		parse_number,    /* 46 */
		parse_slash,     /* 47 */
		parse_number,    /* 48 */
		parse_number,    /* 49 */
		parse_number,    /* 50 */
		parse_number,    /* 51 */
		parse_number,    /* 52 */
		parse_number,    /* 53 */
		parse_number,    /* 54 */
		parse_number,    /* 55 */
		parse_number,    /* 56 */
		parse_number,    /* 57 */
		parse_operator2, /* 58 */
		parse_char,      /* 59 */
		parse_operator2, /* 60 */
		parse_operator2, /* 61 */
		parse_operator2, /* 62 */
		parse_other,     /* 63 */
		parse_var,       /* 64 */
		parse_word,      /* 65 */
		parse_bstring,   /* 66 */
		parse_word,      /* 67 */
		parse_word,      /* 68 */
		parse_estring,   /* 69 */
		parse_word,      /* 70 */
		parse_word,      /* 71 */
		parse_word,      /* 72 */
		parse_word,      /* 73 */
		parse_word,      /* 74 */
		parse_word,      /* 75 */
		parse_word,      /* 76 */
		parse_word,      /* 77 */
		parse_nqstring,  /* 78 */
		parse_word,      /* 79 */
		parse_word,      /* 80 */
		parse_qstring,   /* 81 */
		parse_word,      /* 82 */
		parse_word,      /* 83 */
		parse_word,      /* 84 */
		parse_ustring,   /* 85 */
		parse_word,      /* 86 */
		parse_word,      /* 87 */
		parse_xstring,   /* 88 */
		parse_word,      /* 89 */
		parse_word,      /* 90 */
		parse_bword,     /* 91 */
		parse_backslash, /* 92 */
		parse_other,     /* 93 */
		parse_operator1, /* 94 */
		parse_word,      /* 95 */
		parse_tick,      /* 96 */
		parse_word,      /* 97 */
		parse_bstring,   /* 98 */
		parse_word,      /* 99 */
		parse_word,      /* 100 */
		parse_estring,   /* 101 */
		parse_word,      /* 102 */
		parse_word,      /* 103 */
		parse_word,      /* 104 */
		parse_word,      /* 105 */
		parse_word,      /* 106 */
		parse_word,      /* 107 */
		parse_word,      /* 108 */
		parse_word,      /* 109 */
		parse_nqstring,  /* 110 */
		parse_word,      /* 111 */
		parse_word,      /* 112 */
		parse_qstring,   /* 113 */
		parse_word,      /* 114 */
		parse_word,      /* 115 */
		parse_word,      /* 116 */
		parse_ustring,   /* 117 */
		parse_word,      /* 118 */
		parse_word,      /* 119 */
		parse_xstring,   /* 120 */
		parse_word,      /* 121 */
		parse_word,      /* 122 */
		parse_char,      /* 123 */
		parse_operator2, /* 124 */
		parse_char,      /* 125 */
		parse_operator1, /* 126 */
		parse_white,     /* 127 */
		parse_word,      /* 128 */
		parse_word,      /* 129 */
		parse_word,      /* 130 */
		parse_word,      /* 131 */
		parse_word,      /* 132 */
		parse_word,      /* 133 */
		parse_word,      /* 134 */
		parse_word,      /* 135 */
		parse_word,      /* 136 */
		parse_word,      /* 137 */
		parse_word,      /* 138 */
		parse_word,      /* 139 */
		parse_word,      /* 140 */
		parse_word,      /* 141 */
		parse_word,      /* 142 */
		parse_word,      /* 143 */
		parse_word,      /* 144 */
		parse_word,      /* 145 */
		parse_word,      /* 146 */
		parse_word,      /* 147 */
		parse_word,      /* 148 */
		parse_word,      /* 149 */
		parse_word,      /* 150 */
		parse_word,      /* 151 */
		parse_word,      /* 152 */
		parse_word,      /* 153 */
		parse_word,      /* 154 */
		parse_word,      /* 155 */
		parse_word,      /* 156 */
		parse_word,      /* 157 */
		parse_word,      /* 158 */
		parse_word,      /* 159 */
		parse_white,     /* 160 */
		parse_word,      /* 161 */
		parse_word,      /* 162 */
		parse_word,      /* 163 */
		parse_word,      /* 164 */
		parse_word,      /* 165 */
		parse_word,      /* 166 */
		parse_word,      /* 167 */
		parse_word,      /* 168 */
		parse_word,      /* 169 */
		parse_word,      /* 170 */
		parse_word,      /* 171 */
		parse_word,      /* 172 */
		parse_word,      /* 173 */
		parse_word,      /* 174 */
		parse_word,      /* 175 */
		parse_word,      /* 176 */
		parse_word,      /* 177 */
		parse_word,      /* 178 */
		parse_word,      /* 179 */
		parse_word,      /* 180 */
		parse_word,      /* 181 */
		parse_word,      /* 182 */
		parse_word,      /* 183 */
		parse_word,      /* 184 */
		parse_word,      /* 185 */
		parse_word,      /* 186 */
		parse_word,      /* 187 */
		parse_word,      /* 188 */
		parse_word,      /* 189 */
		parse_word,      /* 190 */
		parse_word,      /* 191 */
		parse_word,      /* 192 */
		parse_word,      /* 193 */
		parse_word,      /* 194 */
		parse_word,      /* 195 */
		parse_word,      /* 196 */
		parse_word,      /* 197 */
		parse_word,      /* 198 */
		parse_word,      /* 199 */
		parse_word,      /* 200 */
		parse_word,      /* 201 */
		parse_word,      /* 202 */
		parse_word,      /* 203 */
		parse_word,      /* 204 */
		parse_word,      /* 205 */
		parse_word,      /* 206 */
		parse_word,      /* 207 */
		parse_word,      /* 208 */
		parse_word,      /* 209 */
		parse_word,      /* 210 */
		parse_word,      /* 211 */
		parse_word,      /* 212 */
		parse_word,      /* 213 */
		parse_word,      /* 214 */
		parse_word,      /* 215 */
		parse_word,      /* 216 */
		parse_word,      /* 217 */
		parse_word,      /* 218 */
		parse_word,      /* 219 */
		parse_word,      /* 220 */
		parse_word,      /* 221 */
		parse_word,      /* 222 */
		parse_word,      /* 223 */
		parse_word,      /* 224 */
		parse_word,      /* 225 */
		parse_word,      /* 226 */
		parse_word,      /* 227 */
		parse_word,      /* 228 */
		parse_word,      /* 229 */
		parse_word,      /* 230 */
		parse_word,      /* 231 */
		parse_word,      /* 232 */
		parse_word,      /* 233 */
		parse_word,      /* 234 */
		parse_word,      /* 235 */
		parse_word,      /* 236 */
		parse_word,      /* 237 */
		parse_word,      /* 238 */
		parse_word,      /* 239 */
		parse_word,      /* 240 */
		parse_word,      /* 241 */
		parse_word,      /* 242 */
		parse_word,      /* 243 */
		parse_word,      /* 244 */
		parse_word,      /* 245 */
		parse_word,      /* 246 */
		parse_word,      /* 247 */
		parse_word,      /* 248 */
		parse_word,      /* 249 */
		parse_word,      /* 250 */
		parse_word,      /* 251 */
		parse_word,      /* 252 */
		parse_word,      /* 253 */
		parse_word,      /* 254 */
		parse_word,      /* 255 */
	}
)

/* Parsers
 *
 * 适当做了部分修改
 */

func parse_white(sf *sqli_state) int {
	return sf.pos + 1
}

func parse_operator1(sf *sqli_state) int {
	pos := sf.pos
	st_assign_char(sf.current, TYPE_OPERATOR, pos, 1, sf.s[pos:pos+1])
	return pos + 1
}

func parse_other(sf *sqli_state) int {
	pos := sf.pos
	st_assign_char(sf.current, TYPE_UNKNOWN, pos, 1, sf.s[pos:pos+1])
	return pos + 1
}

func parse_char(sf *sqli_state) int {
	pos := sf.pos
	st_assign_char(sf.current, sf.s[pos], pos, 1, sf.s[pos:pos+1])
	return pos + 1
}

func parse_eol_comment(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	endpos := memchr(cs, pos, '\n')
	if endpos == -1 {
		st_assign(sf.current, TYPE_COMMENT, pos, slen-pos, cs[pos:])
		return slen
	} else {
		st_assign(sf.current, TYPE_COMMENT, pos, endpos-pos, cs[pos:])
		return endpos + 1
	}
}

/** In ANSI mode, hash is an operator
 *  In MYSQL mode, it's a EOL comment like '--'
 */
func parse_hash(sf *sqli_state) int {
	sf.stats_comment_hash += 1
	if (sf.flags & FLAG_SQL_MYSQL) > 0 {
		sf.stats_comment_hash += 1
		return parse_eol_comment(sf)
	} else {
		st_assign_char(sf.current, TYPE_OPERATOR, sf.pos, 1, "#")
		return sf.pos + 1
	}
}

func parse_dash(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	/*
	 * five cases
	 * 1) --[white]  this is always a SQL comment
	 * 2) --[EOF]    this is a comment
	 * 3) --[notwhite] in MySQL this is NOT a comment but two unary operators
	 * 4) --[notwhite] everyone else thinks this is a comment
	 * 5) -[not dash]  '-' is a unary operator
	 */

	if pos+2 < slen && cs[pos+1] == '-' && char_is_white(cs[pos+2]) {
		return parse_eol_comment(sf)
	} else if pos+2 == slen && cs[pos+1] == '-' {
		return parse_eol_comment(sf)
	} else if pos+1 < slen && cs[pos+1] == '-' && (sf.flags&FLAG_SQL_ANSI) > 0 {
		/* --[not-white] not-white case:
		 *
		 */
		sf.stats_comment_ddx += 1
		return parse_eol_comment(sf)
	} else {
		st_assign_char(sf.current, TYPE_OPERATOR, pos, 1, "-")
		return pos + 1
	}
}

/** This detects MySQL comments, comments that
 * start with /x!   We just ban these now but
 * previously we attempted to parse the inside
 *
 * For reference:
 * the form of /x![anything]x/ or /x!12345[anything] x/
 *
 * Mysql 3 (maybe 4), allowed this:
 *    /x!0selectx/ 1;
 * where 0 could be any number.
 *
 * The last version of MySQL 3 was in 2003.

 * It is unclear if the MySQL 3 syntax was allowed
 * in MySQL 4.  The last version of MySQL 4 was in 2008
 *
 */
func is_mysql_comment(cs string, len, pos int) bool {
	/* so far...
	 * cs[pos] == '/' && cs[pos+1] == '*'
	 */

	if pos+2 >= len {
		/* not a mysql comment */
		return false
	}

	if cs[pos+2] != '!' {
		/* not a mysql comment */
		return false
	}

	/*
	 * this is a mysql comment
	 *  got "/x!"
	 */
	return true
}

func parse_slash(sf *sqli_state) int {
	var clen int
	cs := sf.s
	slen := sf.slen
	pos := sf.pos
	cur := cs[pos:]
	ctype := uint8(TYPE_COMMENT)
	pos1 := pos + 1
	if pos1 == slen || cs[pos1] != '*' {
		return parse_operator1(sf)
	}

	/*
	 * skip over initial '/x'
	 */
	ptr := memchr2(cur[2:], "*/")

	/*
	 * (ptr == NULL) causes false positive in cppcheck 1.61
	 * casting to type seems to fix it
	 */
	if ptr == -1 {
		/* till end of line */
		clen = slen - pos
	} else {
		clen = ptr + 4
	}

	/*
	 * postgresql allows nested comments which makes
	 * this is incompatible with parsing so
	 * if we find a '/x' inside the coment, then
	 * make a new token.
	 *
	 * Also, Mysql's "conditional" comments for version
	 *  are an automatic black ban!
	 */

	if memchr2(cur[2:], "/*") != -1 {
		ctype = TYPE_EVIL
	} else if is_mysql_comment(cs, slen, pos) {
		ctype = TYPE_EVIL
	}

	st_assign(sf.current, ctype, pos, clen, cs[pos:])
	return pos + clen
}

func parse_backslash(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	/*
	 * Weird MySQL alias for NULL, "\N" (capital N only)
	 */
	if pos+1 < slen && cs[pos+1] == 'N' {
		st_assign(sf.current, TYPE_NUMBER, pos, 2, cs[pos:])
		return pos + 2
	} else {
		st_assign_char(sf.current, TYPE_BACKSLASH, pos, 1, cs[pos:pos+1])
		return pos + 1
	}
}

func parse_operator2(sf *sqli_state) int {
	var ch uint8
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	if pos+1 >= slen {
		return parse_operator1(sf)
	}

	if pos+2 < slen &&
		cs[pos] == '<' &&
		cs[pos+1] == '=' &&
		cs[pos+2] == '>' {
		/*
		 * special 3-char operator
		 */
		st_assign(sf.current, TYPE_OPERATOR, pos, 3, cs[pos:])
		return pos + 3
	}

	ch = uint8(sf.lookup(sf, LOOKUP_OPERATOR, []byte(cs[pos:])))
	if ch != 0 {
		st_assign(sf.current, ch, pos, 2, cs[pos:])
		return pos + 2
	}

	/*
	 * not an operator.. what to do with the two
	 * characters we got?
	 */

	if cs[pos] == ':' {
		/* ':' is not an operator */
		st_assign(sf.current, TYPE_COLON, pos, 1, cs[pos:])
		return pos + 1
	} else {
		/*
		 * must be a single char operator
		 */
		return parse_operator1(sf)
	}
}

/*
 * Ok!   "  \"   "  one backslash = escaped!
 *       " \\"   "  two backslash = not escaped!
 *       "\\\"   "  three backslash = escaped!
 */
func is_backslash_escaped(src string, start, end int) bool {

	for i := end; i >= start; i-- {
		if src[i] != '\\' {
			break
		}
	}

	return ((end - start) & 1) > 0
}

func is_double_delim_escaped(src string, cur, end int) bool {
	return (cur+1) < end && src[cur+1] == src[cur]
}

/* Look forward for doubling of delimiter
 *
 * case 'foo''bar' -. foo''bar
 *
 * ending quote isn't duplicated (i.e. escaped)
 * since it's the wrong char or EOL
 *
 */
func parse_string_core(cs string, len, pos int, st *sqli_token, delim byte, offset int) int {
	/*
	 * offset is to skip the perhaps first quote char
	 */
	qpos := memchr(cs, pos+offset, delim)

	/*
	 * then keep string open/close info
	 */
	if offset > 0 {
		/*
		 * this is real quote
		 */
		st.str_open = delim
	} else {
		/*
		 * this was a simulated quote
		 */
		st.str_open = CHAR_NULL
	}

	for {
		if qpos == -1 {
			/*
			 * string ended with no trailing quote
			 * assign what we have
			 */
			st_assign(st, TYPE_STRING, pos+offset, len-pos-offset, cs[pos+offset:])
			st.str_close = CHAR_NULL
			return len

		} else if is_backslash_escaped(cs, qpos-1, pos+offset) {
			/* keep going, move ahead one character */
			qpos = memchr(cs, qpos+1, delim)
			continue
		} else if is_double_delim_escaped(cs, qpos, len) {
			/* keep going, move ahead two characters */
			qpos = memchr(cs, qpos+2, delim)
			continue
		} else {
			/* hey it's a normal string */
			st_assign(st, TYPE_STRING, pos+offset, qpos-(pos+offset), cs[pos+offset:])
			st.str_close = delim
			return qpos + 1
		}
	}
}

/**
 * Used when first char is a ' or "
 */
func parse_string(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	/*
	 * assert cs[pos] == single or double quote
	 */
	return parse_string_core(cs, slen, pos, sf.current, cs[pos], 1)
}

/**
 * Used when first char is:
 *    N or n:  mysql "National Character set"
 *    E     :  psql  "Escaped String"
 */
func parse_estring(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	if pos+2 >= slen || cs[pos+1] != CHAR_SINGLE {
		return parse_word(sf)
	}
	return parse_string_core(cs, slen, pos, sf.current, CHAR_SINGLE, 2)
}

func parse_ustring(sf *sqli_state) int {
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	if pos+2 < slen && cs[pos+1] == '&' && cs[pos+2] == '\'' {
		sf.pos += 2
		pos = parse_string(sf)
		sf.current.str_open = 'u'
		if sf.current.str_close == '\'' {
			sf.current.str_close = 'u'
		}
		return pos
	} else {
		return parse_word(sf)
	}
}

func parse_qstring_core(sf *sqli_state, offset int) int {
	var ch uint8
	var strend int
	cs := sf.s
	slen := sf.slen
	pos := sf.pos + offset

	/* if we are already at end of string..
	   if current char is not q or Q
	   if we don't have 2 more chars
	   if char2 != a single quote
	   then, just treat as word
	*/
	if pos >= slen ||
		(cs[pos] != 'q' && cs[pos] != 'Q') ||
		pos+2 >= slen ||
		cs[pos+1] != '\'' {
		return parse_word(sf)
	}

	ch = cs[pos+2]

	/* the ch > 127 is un-needed since
	 * we assume char is signed
	 */
	if ch < 33 {
		return parse_word(sf)
	}
	switch ch {
	case '(':
		ch = ')'
		break
	case '[':
		ch = ']'
		break
	case '{':
		ch = '}'
		break
	case '<':
		ch = '>'
		break
	}

	strend = memchr2(cs[pos+3:], string([]byte{ch, '\''}))
	if strend == -1 {
		st_assign(sf.current, TYPE_STRING, pos+3, slen-pos-3, cs[pos+3:])
		sf.current.str_open = 'q'
		sf.current.str_close = CHAR_NULL
		return slen
	} else {
		st_assign(sf.current, TYPE_STRING, pos+3, len(cs[pos+3:]), cs[pos+3:])
		sf.current.str_open = 'q'
		sf.current.str_close = 'q'
		return pos + 3 + strend + 2
	}
}

/*
 * Oracle's q string
 */
func parse_qstring(sf *sqli_state) int {
	return parse_qstring_core(sf, 0)
}

/*
 * mysql's N'STRING' or
 * ...  Oracle's nq string
 */
func parse_nqstring(sf *sqli_state) int {
	slen := sf.slen
	pos := sf.pos
	if pos+2 < slen && sf.s[pos+1] == CHAR_SINGLE {
		return parse_estring(sf)
	}
	return parse_qstring_core(sf, 1)
}

/*
 * binary literal string
 * re: [bB]'[01]*'
 */
func parse_bstring(sf *sqli_state) int {
	var wlen int
	cs := sf.s
	pos := sf.pos
	slen := sf.slen

	/* need at least 2 more characters
	 * if next char isn't a single quote, then
	 * continue as normal word
	 */
	if pos+2 >= slen || cs[pos+1] != '\'' {
		return parse_word(sf)
	}

	wlen = strlenspn(cs[pos+2:], "01")
	if pos+2+wlen >= slen || cs[pos+2+wlen] != '\'' {
		return parse_word(sf)
	}
	st_assign(sf.current, TYPE_NUMBER, pos, wlen+3, cs[pos:])
	return pos + 2 + wlen + 1
}

/*
 * hex literal string
 * re: [xX]'[0123456789abcdefABCDEF]*'
 * mysql has requirement of having EVEN number of chars,
 *  but pgsql does not
 */
func parse_xstring(sf *sqli_state) int {
	var wlen int
	cs := sf.s
	pos := sf.pos
	slen := sf.slen

	/* need at least 2 more characters
	 * if next char isn't a single quote, then
	 * continue as normal word
	 */
	if pos+2 >= slen || cs[pos+1] != '\'' {
		return parse_word(sf)
	}

	wlen = strlenspn(cs[pos+2:], "0123456789ABCDEFabcdef")
	if pos+2+wlen >= slen || cs[pos+2+wlen] != '\'' {
		return parse_word(sf)
	}
	st_assign(sf.current, TYPE_NUMBER, pos, wlen+3, cs[pos:])
	return pos + 2 + wlen + 1
}

/**
 * This handles MS SQLSERVER bracket words
 * http://stackoverflow.com/questions/3551284/sql-serverwhat-do-brackets-mean-around-column-name
 *
 */
func parse_bword(sf *sqli_state) int {
	cs := sf.s
	pos := sf.pos
	endptr := memchr(cs, pos, ']')
	if endptr == -1 {
		st_assign(sf.current, TYPE_BAREWORD, pos, sf.slen-pos, cs[pos:])
		return sf.slen
	} else {
		st_assign(sf.current, TYPE_BAREWORD, pos, endptr-pos+1, cs[pos:])
		return endptr + 1
	}
}

func parse_word(sf *sqli_state) int {
	var ch uint8
	var delim byte
	cs := sf.s
	pos := sf.pos
	wlen := strlencspn(cs[pos:], " []{}<>:\\?=@!#~+-*/&|^%(),';\t\n\v\f\r\"\240\000")

	st_assign(sf.current, TYPE_BAREWORD, pos, wlen, cs[pos:])

	/* now we need to look inside what we good for "." and "`"
	 * and see if what is before is a keyword or not
	 */
	for i := 0; i < sf.current.len; i++ {
		delim = sf.current.val[i]
		if delim == '.' || delim == '`' {
			ch = uint8(sf.lookup(sf, LOOKUP_WORD, []byte(sf.current.val)))
			if ch != TYPE_NONE && ch != TYPE_BAREWORD {
				/* needed for swig */
				st_clear(sf.current)
				/*
				 * we got something like "SELECT.1"
				 * or SELECT`column`
				 */
				st_assign(sf.current, ch, pos, i, cs[pos:])
				return pos + i
			}
		}
	}

	/*
	 * do normal lookup with word including '.'
	 */
	if wlen < LIBINJECTION_SQLI_TOKEN_SIZE {

		ch = uint8(sf.lookup(sf, LOOKUP_WORD, []byte(sf.current.val)))
		if ch == CHAR_NULL {
			ch = TYPE_BAREWORD
		}
		sf.current.ttype = ch
	}
	return pos + wlen
}

/* MySQL backticks are a cross between string and
 * and a bare word.
 *
 */
func parse_tick(sf *sqli_state) int {
	pos := parse_string_core(sf.s, sf.slen, sf.pos, sf.current, CHAR_TICK, 1)

	/* we could check to see if start and end of
	 * of string are both "`", i.e. make sure we have
	 * matching set.  `foo` vs. `foo
	 * but I don't think it matters much
	 */

	/* check value of string to see if it's a keyword,
	 * function, operator, etc
	 */
	ch := sf.lookup(sf, LOOKUP_WORD, []byte(sf.current.val))
	if ch == TYPE_FUNCTION {
		/* if it's a function, then convert token */
		sf.current.ttype = TYPE_FUNCTION
	} else {
		/* otherwise it's a 'n' type -- mysql treats
		 * everything as a bare word
		 */
		sf.current.ttype = TYPE_BAREWORD
	}
	return pos
}

func parse_var(sf *sqli_state) int {
	var xlen int
	cs := sf.s
	slen := sf.slen
	pos := sf.pos + 1

	/*
	 * var_count is only used to reconstruct
	 * the input.  It counts the number of '@'
	 * seen 0 in the case of NULL, 1 or 2
	 */

	/*
	 * move past optional other '@'
	 */
	if pos < slen && cs[pos] == '@' {
		pos += 1
		sf.current.count = 2
	} else {
		sf.current.count = 1
	}

	/*
	 * MySQL allows @@`version`
	 */
	if pos < slen {
		if cs[pos] == '`' {
			sf.pos = pos
			pos = parse_tick(sf)
			sf.current.ttype = TYPE_VARIABLE
			return pos
		} else if cs[pos] == CHAR_SINGLE || cs[pos] == CHAR_DOUBLE {
			sf.pos = pos
			pos = parse_string(sf)
			sf.current.ttype = TYPE_VARIABLE
			return pos
		}
	}

	xlen = strlencspn(cs[pos:], " <>:\\?=@!#~+-*/&|^%(),';\t\n\v\f\r'`\"")
	if xlen == 0 {
		st_assign(sf.current, TYPE_VARIABLE, pos, 0, cs[pos:])
		return pos
	} else {
		st_assign(sf.current, TYPE_VARIABLE, pos, xlen, cs[pos:])
		return pos + xlen
	}
}

func parse_money(sf *sqli_state) int {
	var xlen int
	var strend int
	cs := sf.s
	slen := sf.slen
	pos := sf.pos

	if pos+1 == slen {
		/* end of line */
		st_assign_char(sf.current, TYPE_BAREWORD, pos, 1, "$")
		return slen
	}

	/*
	 * $1,000.00 or $1.000,00 ok!
	 * This also parses $....,,,111 but that's ok
	 */

	xlen = strlenspn(cs[pos+1:], "0123456789.,")
	if xlen == 0 {
		if cs[pos+1] == '$' {
			/* we have $$ .. find ending $$ and make string */
			strend = memchr2(cs[pos+2:], "$$")
			if strend == -1 {
				/* fell off edge */
				st_assign(sf.current, TYPE_STRING, pos+2, slen-(pos+2), cs[pos+2:])
				sf.current.str_open = '$'
				sf.current.str_close = CHAR_NULL
				return slen
			} else {
				st_assign(sf.current, TYPE_STRING, pos+2, len(cs[pos+2:]), cs[pos+2:])
				sf.current.str_open = '$'
				sf.current.str_close = '$'
				return strend + 2
			}
		} else {
			/* ok it's not a number or '$$', but maybe it's pgsql "$ quoted strings" */
			xlen = strlenspn(cs[pos+1:], "abcdefghjiklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			if xlen == 0 {
				/* hmm it's "$" _something_ .. just add $ and keep going*/
				st_assign_char(sf.current, TYPE_BAREWORD, pos, 1, "$")
				return pos + 1
			}
			/* we have $foobar????? */
			/* is it $foobar$ */
			if pos+xlen+1 == slen || cs[pos+xlen+1] != '$' {
				/* not $foobar$, or fell off edge */
				st_assign_char(sf.current, TYPE_BAREWORD, pos, 1, "$")
				return pos + 1
			}

			/* we have $foobar$ ... find it again */
			strend = my_memmem(cs[xlen+2:], cs[pos:])

			if strend == -1 || strend < (pos+xlen+2) {
				/* fell off edge */
				st_assign(sf.current, TYPE_STRING, pos+xlen+2, slen-pos-xlen-2, cs[pos+xlen+2:])
				sf.current.str_open = '$'
				sf.current.str_close = CHAR_NULL
				return slen
			} else {
				/* got one */
				st_assign(sf.current, TYPE_STRING, pos+xlen+2, strend-(pos+xlen+2), cs[pos+xlen+2:])
				sf.current.str_open = '$'
				sf.current.str_close = '$'
				return (strend + xlen + 2)
			}
		}
	} else if xlen == 1 && cs[pos+1] == '.' {
		/* $. should parsed as a word */
		return parse_word(sf)
	} else {
		st_assign(sf.current, TYPE_NUMBER, pos, 1+xlen, cs[pos:])
		return pos + 1 + xlen
	}
}

func parse_number(sf *sqli_state) int {
	var xlen int
	var start int
	var digits string
	cs := sf.s
	slen := sf.slen
	pos := sf.pos
	have_e := 0
	have_exp := 0

	/* cs[pos] == '0' has 1/10 chance of being true,
	 * while pos+1< slen is almost always true
	 */
	if cs[pos] == '0' && pos+1 < slen {
		if cs[pos+1] == 'X' || cs[pos+1] == 'x' {
			digits = "0123456789ABCDEFabcdef"
		} else if cs[pos+1] == 'B' || cs[pos+1] == 'b' {
			digits = "01"
		}

		if len(digits) > 0 {
			xlen = strlenspn(cs[pos+2:], digits)
			if xlen == 0 {
				st_assign(sf.current, TYPE_BAREWORD, pos, 2, cs[pos:])
				return pos + 2
			} else {
				st_assign(sf.current, TYPE_NUMBER, pos, 2+xlen, cs[pos:])
				return pos + 2 + xlen
			}
		}
	}

	start = pos
	for pos < slen && ISDIGIT(cs[pos]) {
		pos += 1
	}

	if pos < slen && cs[pos] == '.' {
		pos += 1
		for pos < slen && ISDIGIT(cs[pos]) {
			pos += 1
		}
		if pos-start == 1 {
			/* only one character read so far */
			st_assign_char(sf.current, TYPE_DOT, start, 1, ".")
			return pos
		}
	}

	if pos < slen {
		if cs[pos] == 'E' || cs[pos] == 'e' {
			have_e = 1
			pos += 1
			if pos < slen && (cs[pos] == '+' || cs[pos] == '-') {
				pos += 1
			}
			for pos < slen && ISDIGIT(cs[pos]) {
				have_exp = 1
				pos += 1
			}
		}
	}

	/* oracle's ending float or double suffix
	 * http://docs.oracle.com/cd/B19306_01/server.102/b14200/sql_elements003.htm#i139891
	 */
	if pos < slen && (cs[pos] == 'd' || cs[pos] == 'D' || cs[pos] == 'f' || cs[pos] == 'F') {
		if pos+1 == slen {
			/* line ends evaluate "... 1.2f$" as '1.2f' */
			pos += 1
		} else if char_is_white(cs[pos+1]) || cs[pos+1] == ';' {
			/*
			 * easy case, evaluate "... 1.2f ... as '1.2f'
			 */
			pos += 1
		} else if cs[pos+1] == 'u' || cs[pos+1] == 'U' {
			/*
			 * a bit of a hack but makes '1fUNION' parse as '1f UNION'
			 */
			pos += 1
		} else {
			/* it's like "123FROM" */
			/* parse as "123" only */
		}
	}

	if have_e == 1 && have_exp == 0 {
		/* very special form of
		 * "1234.e"
		 * "10.10E"
		 * ".E"
		 * this is a WORD not a number!! */
		st_assign(sf.current, TYPE_BAREWORD, start, pos-start, cs[start:])
	} else {
		st_assign(sf.current, TYPE_NUMBER, start, pos-start, cs[start:])
	}
	return pos
}
