package GSQLI

import (
	"strings"
)

func h5_is_white(ch uint8) bool {
	/*
	 * \t = horizontal tab = 0x09
	 * \n = newline = 0x0A
	 * \v = vertical tab = 0x0B
	 * \f = form feed = 0x0C
	 * \r = cr  = 0x0D
	 */
	return (ch == '\t' || ch == '\n' || ch == '\v' || ch == '\f' || ch == '\r' || ch == ' ')
}

func memchr(src string, pos int, char byte) int {
	ipos := strings.IndexByte(src[pos:], char)
	if ipos == -1 {
		return -1
	}

	return ipos + pos
}

/* memchr2 finds a string of 2 characters inside another string
 * This a specialized version of "memmem" or "memchr".
 * 'memmem' doesn't exist on all platforms
 *
 * Porting notes: this is just a special version of
 *    astring.find("AB")
 *
 */
func memchr2(haystack string, c0 string) int {
	if len(haystack) < 2 {
		return -1
	}

	return strings.Index(haystack, c0)
}

/**
 * memmem might not exist on some systems
 */
func my_memmem(haystack string, needle string) int {
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle[0] && strings.Index(haystack[i:], needle) != -1 {
			return i
		}
	}
	return -1
}

/** Find largest string containing certain characters.
 *
 * C Standard library 'strspn' only works for 'c-strings' (null terminated)
 * This works on arbitrary length.
 *
 * Performance notes:
 *   not critical
 *
 * Porting notes:
 *   if accept is 'ABC', then this function would be similar to
 *   a_regexp.match(a_str, '[ABC]*'),
 */

func strlenspn(s string, accept string) int {
	for i := 0; i < len(s); i++ {
		/* likely we can do better by inlining this function
		 * but this works for now
		 */
		if strchr(accept, s[i]) == -1 {
			return i
		}
	}
	return len(s)
}

func strlencspn(s string, accept string) int {
	for i := 0; i < len(s); i++ {
		/* likely we can do better by inlining this function
		 * but this works for now
		 */
		if strchr(accept, s[i]) != -1 {
			return i
		}
	}
	return len(s)
}

func strchr(s string, ch uint8) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ch {
			return i
		}
	}

	return -1
}

func char_is_white(ch uint8) bool {
	/* ' '  space is 0x32
	'\t  0x09 \011 horizontal tab
	'\n' 0x0a \012 new line
	'\v' 0x0b \013 vertical tab
	'\f' 0x0c \014 new page
	'\r' 0x0d \015 carriage return
		 0x00 \000 null (oracle)
		 0xa0 \240 is Latin-1
	*/
	return strchr(" \t\n\v\f\r\240\000", ch) != -1
}

/* DANGER DANGER
 * This is -very specialized function-
 *
 * this compares a ALL_UPPER CASE C STRING
 * with a *arbitrary memory* + length
 *
 * Sane people would just make a copy, up-case
 * and use a hash table.
 *
 * Required since libc version uses the current locale
 * and is much slower.
 */
func cstrcasecmp(a string, b string) int {
	return strings.Compare(strings.ToUpper(a), strings.ToUpper(b))
}

/**
 * Case sensitive string compare.
 *  Here only to make code more readable
 */
func streqp(a string, b string) int {
	return strings.Compare(a, b)
}

// 使用map进行搜索，比bsearch还快
func bsearch_keyword_type(key string) uint8 {
	kt, has := szKeywordMap[strings.ToUpper(key)]
	if has {
		return kt.vtype
	}

	return 0
}

func is_keyword(key string) bool {
	return bsearch_keyword_type(key) > 0
}

func st_clear(s *sqli_token) {
	s.count = 0
	s.len = 0
	s.pos = 0
	s.ttype = 0
	s.val = ""
}

func st_assign_char(s *sqli_token, stype uint8, pos int, len int, value string) {
	/* done to eliminate unused warning */
	s.ttype = stype
	s.pos = pos
	s.len = 1
	s.val = value
}

func st_assign(st *sqli_token, stype uint8, pos int, len int, value string) {
	MSIZE := 32
	last := len
	if last >= MSIZE {
		last = MSIZE - 1
	}

	st.ttype = stype
	st.pos = pos
	st.len = last
	st.val = value[:last]
}

func st_copy(dst, src *sqli_token) {
	//reflect.Copy(reflect.ValueOf(dst), reflect.ValueOf(src))
	*dst = *src

	/*dst.ttype = src.ttype
	dst.val = src.val
	dst.count = src.count
	dst.len = src.len
	dst.pos = src.pos
	dst.str_close = src.str_close
	dst.str_open = src.str_open*/
}

func st_is_arithmetic_op(st *sqli_token) bool {
	ch := st.val[0]
	return (st.ttype == TYPE_OPERATOR && st.len == 1 &&
		(ch == '*' || ch == '/' || ch == '-' || ch == '+' || ch == '%'))
}

func st_is_unary_op(st *sqli_token) bool {
	str := st.val
	len := st.len

	if st.ttype != TYPE_OPERATOR {
		return false
	}

	switch len {
	case 1:
		return str[0] == '+' || str[0] == '-' || str[0] == '!' || str[0] == '~'
	case 2:
		return str[0] == '!' && str[1] == '!'
	case 3:
		return cstrcasecmp("NOT", str[:3]) == 0
	default:
		return false
	}

	return false
}

func streq(a, b string) bool {
	return strings.Compare(a, b) == 0
}
