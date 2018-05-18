// from libinjection c code
// libinjection Copyright (c) 2012-2016 Nick Galbreath
// golang author: koangel(jackliu100@gmail.com)

package GSQLI

import (
	"strings"
)

const (
	TYPE_NONE     = iota
	TYPE_BLACK    /* ban always */
	TYPE_ATTR_URL /* attribute value takes a URL-like object */
	TYPE_STYLE
	TYPE_ATTR_INDIRECT /* attribute *name* is given in *value* */
)

type stringtype_t struct {
	name  string
	atype int
}

var gsHexDecodeMap = [256]int{
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 256, 256,
	256, 256, 256, 256, 256, 10, 11, 12, 13, 14, 15, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 10, 11, 12, 13, 14, 15, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256,
	256, 256, 256, 256,
}

func html_decode_char_at(src string) (ret, consumed int) {
	var val int = 0
	var i int
	var ch int

	if len(src) == 0 {
		consumed = 0
		ret = -1
		return
	}

	consumed = 1
	if src[0] != '&' || len(src) < 2 {
		ret = int(src[0])
		return
	}

	if src[1] != '#' {
		/* normally this would be for named entities
		 * but for this case we don't actually care
		 */
		ret = int('&')
		return
	}

	if src[2] == 'x' || src[2] == 'X' {
		ch = int(src[3])
		ch = gsHexDecodeMap[ch]
		if int(ch) == 256 {
			/* degenerate case  '&#[?]' */
			ret = int('&')
			return
		}
		val = int(ch)
		i = 4
		for i < len(src) {
			ch = int(src[i])
			if ch == ';' {
				consumed = i + 1
				ret = val
				return
			}
			ch = gsHexDecodeMap[ch]
			if int(ch) == 256 {
				consumed = i
				ret = val
				return
			}

			val = (val * 16) + int(ch)
			if val > 0x1000FF {
				ret = int('&')
				return
			}
			i++
		}
		consumed = i
		ret = val
		return
	} else {
		i = 2
		ch = int(src[i])
		if ch < '0' || ch > '9' {
			ret = int('&')
			return
		}
		val = int(ch) - '0'
		i += 1
		for i < len(src) {
			ch = int(src[i])
			if ch == ';' {
				consumed = i + 1
				ret = val
				return
			}
			if ch < '0' || ch > '9' {
				consumed = i
				ret = val
				return
			}
			val = (val * 10) + int(ch-'0')
			if val > 0x1000FF {
				ret = int('&')
				return
			}
			i++
		}
		consumed = i
		ret = val
		return
	}
}

/*
 * view-source:
 * data:
 * javascript:
 */
var BLACKATTR = []stringtype_t{
	{"ACTION", TYPE_ATTR_URL},             /* form */
	{"ATTRIBUTENAME", TYPE_ATTR_INDIRECT}, /* SVG allow indirection of attribute names */
	{"BY", TYPE_ATTR_URL},                 /* SVG */
	{"BACKGROUND", TYPE_ATTR_URL},         /* IE6, O11 */
	{"DATAFORMATAS", TYPE_BLACK},          /* IE */
	{"DATASRC", TYPE_BLACK},               /* IE */
	{"DYNSRC", TYPE_ATTR_URL},             /* Obsolete img attribute */
	{"FILTER", TYPE_STYLE},                /* Opera, SVG inline style */
	{"FORMACTION", TYPE_ATTR_URL},         /* HTML 5 */
	{"FOLDER", TYPE_ATTR_URL},             /* Only on A tags, IE-only */
	{"FROM", TYPE_ATTR_URL},               /* SVG */
	{"HANDLER", TYPE_ATTR_URL},            /* SVG Tiny, Opera */
	{"HREF", TYPE_ATTR_URL},
	{"LOWSRC", TYPE_ATTR_URL}, /* Obsolete img attribute */
	{"POSTER", TYPE_ATTR_URL}, /* Opera 10,11 */
	{"SRC", TYPE_ATTR_URL},
	{"STYLE", TYPE_STYLE},
	{"TO", TYPE_ATTR_URL},     /* SVG */
	{"VALUES", TYPE_ATTR_URL}, /* SVG */
	{"XLINK:HREF", TYPE_ATTR_URL},
}

/* xmlns */
/* `xml-stylesheet` > <eval>, <if expr=> */

/*
  static const char* BLACKATTR[] = {
  "ATTRIBUTENAME",
  "BACKGROUND",
  "DATAFORMATAS",
  "HREF",
  "SCROLL",
  "SRC",
  "STYLE",
  "SRCDOC",
  NULL
  };
*/

var BLACKTAG = []string{
	"APPLET",
	/*    , "AUDIO" */
	"BASE",
	"COMMENT", /* IE http://html5sec.org/#38 */
	"EMBED",
	/*   ,  "FORM" */
	"FRAME",
	"FRAMESET",
	"HANDLER", /* Opera SVG, effectively a script tag */
	"IFRAME",
	"IMPORT",
	"ISINDEX",
	"LINK",
	"LISTENER",
	/*    , "MARQUEE" */
	"META",
	"NOSCRIPT",
	"OBJECT",
	"SCRIPT",
	"STYLE",
	/*    , "VIDEO" */
	"VMLFRAME",
	"XML",
	"XSS",
}

func cstrcasecmp_with_null(a, b string, cmplen int) int {
	if strings.Contains(strings.ToUpper(b[:cmplen]), a) {
		return 0
	}

	return 1
}

/*
 * Does an HTML encoded  binary string (const char*, length) start with
 * a all uppercase c-string (null terminated), case insensitive!
 *
 * also ignore any embedded nulls in the HTML string!
 *
 * return 1 if match / starts with
 * return 0 if not
 */
func htmlencode_startswith(a, b string) bool {
	var consumed int
	var cb int
	var first int = 1
	n := len(a)

	/* printf("Comparing %s with %.*s\n", a,(int)n,b); */
	ai := 0
	bi := 0
	for n > 0 {
		if ai >= len(a) {
			/* printf("Match EOL!\n"); */
			return true
		}
		cb, consumed = html_decode_char_at(b[bi:])
		bi += consumed
		n -= consumed

		if first == 1 && cb <= 32 {
			/* ignore all leading whitespace and control characters */
			continue
		}
		first = 0

		if cb == 0 {
			/* always ignore null characters in user input */
			continue
		}

		if cb == 10 {
			/* always ignore vertical tab characters in user input */
			/* who allows this?? */
			continue
		}

		if cb >= 'a' && cb <= 'z' {
			/* upcase */
			cb -= 0x20
		}

		if a[ai] != uint8(cb) {
			/* printf("    %c != %c\n", *a, cb); */
			/* mismatch */
			return false
		}
		ai++
	}

	if ai >= len(a) {
		return true
	}

	return false
}

func is_black_tag(s string, len int) bool {
	if len < 3 {
		return false
	}

	for _, black := range BLACKTAG {
		if cstrcasecmp_with_null(black, s, len) == 0 {
			/* printf("Got black tag %s\n", *black); */
			return true
		}
	}

	/* anything SVG related */
	if (s[0] == 's' || s[0] == 'S') &&
		(s[1] == 'v' || s[1] == 'V') &&
		(s[2] == 'g' || s[2] == 'G') {
		/*        printf("Got SVG tag \n"); */
		return true
	}

	/* Anything XSL(t) related */
	if (s[0] == 'x' || s[0] == 'X') &&
		(s[1] == 's' || s[1] == 'S') &&
		(s[2] == 'l' || s[2] == 'L') {
		/*      printf("Got XSL tag\n"); */
		return true
	}

	return false
}

func is_black_attr(s string, len int) int {
	if len < 2 {
		return TYPE_NONE
	}

	if len >= 5 {
		/* JavaScript on.* */
		if (s[0] == 'o' || s[0] == 'O') && (s[1] == 'n' || s[1] == 'N') {
			/* printf("Got JavaScript on- attribute name\n"); */
			return TYPE_BLACK
		}

		/* XMLNS can be used to create arbitrary tags */
		if cstrcasecmp_with_null("XMLNS", s, len) == 0 || cstrcasecmp_with_null("XLINK", s, len) == 0 {
			/*      printf("Got XMLNS and XLINK tags\n"); */
			return TYPE_BLACK
		}
	}

	for _, black := range BLACKATTR {
		if cstrcasecmp_with_null(black.name, s, len) == 0 {
			/*      printf("Got banned attribute name %s\n", black.name); */
			return black.atype
		}
	}

	return TYPE_NONE
}

func is_black_url(s string) bool {

	data_url := "DATA"
	viewsource_url := "VIEW-SOURCE"

	/* obsolete but interesting signal */
	vbscript_url := "VBSCRIPT"

	/* covers JAVA, JAVASCRIPT, + colon */
	javascript_url := "JAVA"

	/* skip whitespace */
	s = strings.TrimSpace(s)

	if htmlencode_startswith(data_url, s) {
		return true
	}

	if htmlencode_startswith(viewsource_url, s) {
		return true
	}

	if htmlencode_startswith(javascript_url, s) {
		return true
	}

	if htmlencode_startswith(vbscript_url, s) {
		return true
	}
	return false
}

func libinjection_is_xss(s string, flags int) int {
	var h5 h5_state_t
	attr := TYPE_NONE

	libinjection_h5_init(&h5, s, flags)
	for libinjection_h5_next(&h5) == 1 {
		if h5.token_type != ATTR_VALUE {
			attr = TYPE_NONE
		}

		if h5.token_type == DOCTYPE {
			return 1
		} else if h5.token_type == TAG_NAME_OPEN {
			if is_black_tag(h5.s[h5.token_start:], h5.token_len) {
				return 1
			}
		} else if h5.token_type == ATTR_NAME {
			attr = is_black_attr(h5.s[h5.token_start:], h5.token_len)
		} else if h5.token_type == ATTR_VALUE {
			/*
			 * IE6,7,8 parsing works a bit differently so
			 * a whole <script> or other black tag might be hiding
			 * inside an attribute value under HTML 5 parsing
			 * See http://html5sec.org/#102
			 * to avoid doing a full reparse of the value, just
			 * look for "<".  This probably need adjusting to
			 * handle escaped characters
			 */
			/*
			   if (memchr(h5.token_start, '<', h5.token_len) != NULL) {
			   return 1;
			   }
			*/

			switch attr {
			case TYPE_NONE:
				break
			case TYPE_BLACK:
				return 1
			case TYPE_ATTR_URL:
				if is_black_url(h5.s[h5.token_start : h5.token_start+h5.token_len]) {
					return 1
				}
				break
			case TYPE_STYLE:
				return 1
			case TYPE_ATTR_INDIRECT:
				/* an attribute name is specified in a _value_ */
				if is_black_attr(h5.s[h5.token_start:], h5.token_len) > 0 {
					return 1
				}
				break
				/*
				   default:
				   assert(0);
				*/
			}
			attr = TYPE_NONE
		} else if h5.token_type == TAG_COMMENT {
			/* IE uses a "`" as a tag ending char */
			if memchr(h5.s[h5.token_start:], h5.token_start, '`') != 0 {
				return 1
			}

			/* IE conditional comment */
			if h5.token_len > 3 {
				if h5.s[h5.token_start] == '[' &&
					(h5.s[h5.token_start+1] == 'i' || h5.s[h5.token_start+1] == 'I') &&
					(h5.s[h5.token_start+2] == 'f' || h5.s[h5.token_start+2] == 'F') {
					return 1
				}
				if (h5.s[h5.token_start] == 'x' || h5.s[h5.token_start] == 'X') &&
					(h5.s[h5.token_start+1] == 'm' || h5.s[h5.token_start+1] == 'M') &&
					(h5.s[h5.token_start+2] == 'l' || h5.s[h5.token_start+2] == 'L') {
					return 1
				}
			}

			if h5.token_len > 5 {
				/*  IE <?import pseudo-tag */
				if cstrcasecmp_with_null("IMPORT", h5.s[h5.token_start:], h5.token_len) == 0 {
					return 1
				}

				/*  XML Entity definition */
				if cstrcasecmp_with_null("ENTITY", h5.s[h5.token_start:], h5.token_len) == 0 {
					return 1
				}
			}
		}
	}
	return 0
}

func XSSParser(s string) bool {
	if libinjection_is_xss(s, DATA_STATE) == 1 {
		return true
	}
	if libinjection_is_xss(s, VALUE_NO_QUOTE) == 1 {
		return true
	}
	if libinjection_is_xss(s, VALUE_SINGLE_QUOTE) == 1 {
		return true
	}
	if libinjection_is_xss(s, VALUE_DOUBLE_QUOTE) == 1 {
		return true
	}
	if libinjection_is_xss(s, VALUE_BACK_QUOTE) == 1 {
		return true
	}

	return false
}
