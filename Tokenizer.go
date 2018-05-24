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
