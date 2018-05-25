package GSQLI

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

	ttype     uint8
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
		parse_white, /* 0 */
		parse_white, /* 1 */
		parse_white, /* 2 */
		parse_white, /* 3 */
		parse_white, /* 4 */
		parse_white, /* 5 */
		parse_white, /* 6 */
		parse_white, /* 7 */
		parse_white, /* 8 */
		parse_white, /* 9 */
		parse_white, /* 10 */
		parse_white, /* 11 */
		parse_white, /* 12 */
		parse_white, /* 13 */
		parse_white, /* 14 */
		parse_white, /* 15 */
		parse_white, /* 16 */
		parse_white, /* 17 */
		parse_white, /* 18 */
		parse_white, /* 19 */
		parse_white, /* 20 */
		parse_white, /* 21 */
		parse_white, /* 22 */
		parse_white, /* 23 */
		parse_white, /* 24 */
		parse_white, /* 25 */
		parse_white, /* 26 */
		parse_white, /* 27 */
		parse_white, /* 28 */
		parse_white, /* 29 */
		parse_white, /* 30 */
		parse_white, /* 31 */
		parse_white, /* 32 */
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

	endpos := memchr(cs[pos:], pos, '\n')
	if endpos == -1 {
		st_assign(sf.current, TYPE_COMMENT, pos, slen-pos, cs[pos:])
		return slen
	} else {
		st_assign(sf.current, TYPE_COMMENT, pos, endpos-pos, cs[pos:endpos])
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
