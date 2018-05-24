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
	ipos := strings.IndexByte(src, char)
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
	kt, has := fingerprints.Keywords[strings.ToUpper(key)]
	if has {
		return kt[0]
	}

	return 0
}

func is_keyword(key string) bool {
	return bsearch_keyword_type(key) > 0
}
