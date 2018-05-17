// from libinjection c code
// libinjection Copyright (c) 2012-2016 Nick Galbreath
// golang author: koangel(jackliu100@gmail.com)

package GSQLI

import (
	"strings"
)

const (
	DATA_TEXT = iota
	TAG_NAME_OPEN
	TAG_NAME_CLOSE
	TAG_NAME_SELFCLOSE
	TAG_DATA
	TAG_CLOSE
	ATTR_NAME
	ATTR_VALUE
	TAG_COMMENT
	DOCTYPE
)

const (
	CHAR_EOF      = -1
	CHAR_NULL     = 0
	CHAR_BANG     = 33
	CHAR_DOUBLE   = 34
	CHAR_PERCENT  = 37
	CHAR_SINGLE   = 39
	CHAR_DASH     = 45
	CHAR_SLASH    = 47
	CHAR_LT       = 60
	CHAR_EQUALS   = 61
	CHAR_GT       = 62
	CHAR_QUESTION = 63
	CHAR_RIGHTB   = 93
	CHAR_TICK     = 96
)

const (
	DATA_STATE = iota
	VALUE_NO_QUOTE
	VALUE_SINGLE_QUOTE
	VALUE_DOUBLE_QUOTE
	VALUE_BACK_QUOTE
)

type h5parserFn func(src *h5_state_t) int

type h5_state_t struct {
	s           string
	len         int
	pos         int
	is_close    bool
	state       h5parserFn
	token_start int // use pos
	token_len   int
	token_type  int
}

/**
 * public function
 */
func libinjection_h5_init(hs *h5_state_t, s string, flags int) {
	hs.s = s
	hs.len = len(s)
	hs.token_len = 0
	hs.token_start = 0

	switch flags {
	case DATA_STATE:
		hs.state = h5_state_data
		break
	case VALUE_NO_QUOTE:
		hs.state = h5_state_before_attribute_name
		break
	case VALUE_SINGLE_QUOTE:
		hs.state = h5_state_attribute_value_single_quote
		break
	case VALUE_DOUBLE_QUOTE:
		hs.state = h5_state_attribute_value_double_quote
		break
	case VALUE_BACK_QUOTE:
		hs.state = h5_state_attribute_value_back_quote
		break
	}
}

/**
 * public function
 */
func libinjection_h5_next(hs *h5_state_t) int {
	if hs.state != nil {
		return hs.state(hs)
	}
	return 0
}

func h5_is_white(ch uint8) bool {
	/*
	 * \t = horizontal tab = 0x09
	 * \n = newline = 0x0A
	 * \v = vertical tab = 0x0B
	 * \f = form feed = 0x0C
	 * \r = cr  = 0x0D
	 */
	return (ch == '\t' || ch == '\n' || ch == '\v' || ch == '\f' || ch == '\r')
}

func memchr(src string, pos int, char byte) int {
	ipos := strings.IndexByte(src, char)
	if ipos == -1 {
		return -1
	}

	return ipos + pos
}

func h5_skip_white(hs *h5_state_t) int {
	var ch uint8
	for hs.pos < hs.len {
		ch = hs.s[hs.pos]
		switch ch {
		case 0x00, 0x20, 0x09, 0x0A, 0x0B, 0x0C, 0x0D: /* IE only */
			hs.pos += 1
		default:
			return int(ch)
		}
	}
	return CHAR_EOF
}

func h5_state_eof(hs *h5_state_t) int {
	return 0
}

func h5_state_data(hs *h5_state_t) int {
	idx := memchr(hs.s[hs.pos:], hs.pos, CHAR_LT)
	if idx == -1 {
		hs.token_start = hs.pos
		hs.token_len = hs.len - hs.pos
		hs.token_type = DATA_TEXT
		hs.state = h5_state_eof
		if hs.token_len == 0 {
			return 0
		}
	} else {
		hs.token_start = hs.pos
		hs.token_type = DATA_TEXT
		hs.token_len = hs.len - idx - hs.pos
		hs.pos = idx + 1
		hs.state = h5_state_tag_open
		if hs.token_len == 0 {
			return h5_state_tag_open(hs)
		}
	}
	return 1
}

/**
 * 12 2.4.8
 */
func h5_state_tag_open(hs *h5_state_t) int {
	var ch uint8

	if hs.pos >= hs.len {
		return 0
	}
	ch = hs.s[hs.pos]
	if ch == CHAR_BANG {
		hs.pos += 1
		return h5_state_markup_declaration_open(hs)
	} else if ch == CHAR_SLASH {
		hs.pos += 1
		hs.is_close = true
		return h5_state_end_tag_open(hs)
	} else if ch == CHAR_QUESTION {
		hs.pos += 1
		return h5_state_bogus_comment(hs)
	} else if ch == CHAR_PERCENT {
		/* this is not in spec.. alternative comment format used
		   by IE <= 9 and Safari < 4.0.3 */
		hs.pos += 1
		return h5_state_bogus_comment2(hs)
	} else if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
		return h5_state_tag_name(hs)
	} else if ch == CHAR_NULL {
		/* IE-ism  NULL characters are ignored */
		return h5_state_tag_name(hs)
	} else {
		/* user input mistake in configuring state */
		if hs.pos == 0 {
			return h5_state_data(hs)
		}
		hs.token_start = hs.pos - 1
		hs.token_len = 1
		hs.token_type = DATA_TEXT
		hs.state = h5_state_data
		return 1
	}
}

/**
 * 12.2.4.9
 */
func h5_state_end_tag_open(hs *h5_state_t) int {
	var ch uint8

	if hs.pos >= hs.len {
		return 0
	}
	ch = hs.s[hs.pos]
	if ch == CHAR_GT {
		return h5_state_data(hs)
	} else if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
		return h5_state_tag_name(hs)
	}

	hs.is_close = false
	return h5_state_bogus_comment(hs)
}

/*
 *
 */
func h5_state_tag_name_close(hs *h5_state_t) int {

	hs.is_close = false
	hs.token_start = hs.pos
	hs.token_len = 1
	hs.token_type = TAG_NAME_CLOSE
	hs.pos += 1
	if hs.pos < hs.len {
		hs.state = h5_state_data
	} else {
		hs.state = h5_state_eof
	}

	return 1
}

/**
 * 12.2.4.10
 */
func h5_state_tag_name(hs *h5_state_t) int {
	var ch uint8
	var pos int

	pos = hs.pos
	for pos < hs.len {
		ch = hs.s[pos]
		if ch == 0 {
			/* special non-standard case */
			/* allow nulls in tag name   */
			/* some old browsers apparently allow and ignore them */
			pos += 1
		} else if h5_is_white(ch) {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = TAG_NAME_OPEN
			hs.pos = pos + 1
			hs.state = h5_state_before_attribute_name
			return 1
		} else if ch == CHAR_SLASH {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = TAG_NAME_OPEN
			hs.pos = pos + 1
			hs.state = h5_state_self_closing_start_tag
			return 1
		} else if ch == CHAR_GT {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			if hs.is_close {
				hs.pos = pos + 1
				hs.is_close = false
				hs.token_type = TAG_CLOSE
				hs.state = h5_state_data
			} else {
				hs.pos = pos
				hs.token_type = TAG_NAME_OPEN
				hs.state = h5_state_tag_name_close
			}
			return 1
		} else {
			pos += 1
		}
	}

	hs.token_start = hs.pos
	hs.token_len = hs.len - hs.pos
	hs.token_type = TAG_NAME_OPEN
	hs.state = h5_state_eof
	return 1
}

/**
 * 12.2.4.34
 */
func h5_state_before_attribute_name(hs *h5_state_t) int {
	var ch int

	ch = h5_skip_white(hs)
	switch ch {
	case CHAR_EOF:
		{
			return 0
		}
	case CHAR_SLASH:
		{
			hs.pos += 1
			return h5_state_self_closing_start_tag(hs)
		}
	case CHAR_GT:
		{
			hs.state = h5_state_data
			hs.token_start = hs.pos
			hs.token_len = 1
			hs.token_type = TAG_NAME_CLOSE
			hs.pos += 1
			return 1
		}
	default:
		{
			return h5_state_attribute_name(hs)
		}
	}
}

func h5_state_attribute_name(hs *h5_state_t) int {
	var ch uint8
	var pos int

	pos = hs.pos + 1
	for pos < hs.len {
		ch = hs.s[pos]
		if h5_is_white(ch) {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = ATTR_NAME
			hs.state = h5_state_after_attribute_name
			hs.pos = pos + 1
			return 1
		} else if ch == CHAR_SLASH {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = ATTR_NAME
			hs.state = h5_state_self_closing_start_tag
			hs.pos = pos + 1
			return 1
		} else if ch == CHAR_EQUALS {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = ATTR_NAME
			hs.state = h5_state_before_attribute_value
			hs.pos = pos + 1
			return 1
		} else if ch == CHAR_GT {
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.token_type = ATTR_NAME
			hs.state = h5_state_tag_name_close
			hs.pos = pos
			return 1
		} else {
			pos += 1
		}
	}
	/* EOF */
	hs.token_start = hs.pos
	hs.token_len = hs.len - hs.pos
	hs.token_type = ATTR_NAME
	hs.state = h5_state_eof
	hs.pos = hs.len
	return 1
}

/**
 * 12.2.4.36
 */
func h5_state_after_attribute_name(hs *h5_state_t) int {
	var c int

	c = h5_skip_white(hs)
	switch c {
	case CHAR_EOF:
		{
			return 0
		}
	case CHAR_SLASH:
		{
			hs.pos += 1
			return h5_state_self_closing_start_tag(hs)
		}
	case CHAR_EQUALS:
		{
			hs.pos += 1
			return h5_state_before_attribute_value(hs)
		}
	case CHAR_GT:
		{
			return h5_state_tag_name_close(hs)
		}
	default:
		{
			return h5_state_attribute_name(hs)
		}
	}
}

/**
 * 12.2.4.37
 */
func h5_state_before_attribute_value(hs *h5_state_t) int {
	var c int

	c = h5_skip_white(hs)

	if c == CHAR_EOF {
		hs.state = h5_state_eof
		return 0
	}

	if c == CHAR_DOUBLE {
		return h5_state_attribute_value_double_quote(hs)
	} else if c == CHAR_SINGLE {
		return h5_state_attribute_value_single_quote(hs)
	} else if c == CHAR_TICK {
		/* NON STANDARD IE */
		return h5_state_attribute_value_back_quote(hs)
	} else {
		return h5_state_attribute_value_no_quote(hs)
	}
}

func h5_state_attribute_value_quote(hs *h5_state_t, qchar byte) int {
	/* skip initial quote in normal case.
	 * don't do this "if (pos == 0)" since it means we have started
	 * in a non-data state.  given an input of '><foo
	 * we want to make 0-length attribute name
	 */
	if hs.pos > 0 {
		hs.pos += 1
	}

	idx := memchr(hs.s[hs.pos:], hs.pos, qchar)
	if idx == -1 {
		hs.token_start = hs.pos
		hs.token_len = hs.len - hs.pos
		hs.token_type = ATTR_VALUE
		hs.state = h5_state_eof
	} else {
		hs.token_start = hs.pos
		hs.token_len = (hs.len - idx) - hs.pos
		hs.token_type = ATTR_VALUE
		hs.state = h5_state_after_attribute_value_quoted_state
		hs.pos += hs.token_len + 1
	}
	return 1
}

func h5_state_attribute_value_double_quote(hs *h5_state_t) int {
	return h5_state_attribute_value_quote(hs, CHAR_DOUBLE)
}

func h5_state_attribute_value_single_quote(hs *h5_state_t) int {
	return h5_state_attribute_value_quote(hs, CHAR_SINGLE)
}

func h5_state_attribute_value_back_quote(hs *h5_state_t) int {
	return h5_state_attribute_value_quote(hs, CHAR_TICK)
}

func h5_state_attribute_value_no_quote(hs *h5_state_t) int {
	var ch uint8
	var pos int

	pos = hs.pos
	for pos < hs.len {
		ch = hs.s[pos]
		if h5_is_white(ch) {
			hs.token_type = ATTR_VALUE
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.pos = pos + 1
			hs.state = h5_state_before_attribute_name
			return 1
		} else if ch == CHAR_GT {
			hs.token_type = ATTR_VALUE
			hs.token_start = hs.pos
			hs.token_len = pos - hs.pos
			hs.pos = pos
			hs.state = h5_state_tag_name_close
			return 1
		}
		pos += 1
	}

	/* EOF */
	hs.state = h5_state_eof
	hs.token_start = hs.pos
	hs.token_len = hs.len - hs.pos
	hs.token_type = ATTR_VALUE
	return 1
}

/**
 * 12.2.4.41
 */
func h5_state_after_attribute_value_quoted_state(hs *h5_state_t) int {
	var ch uint8

	if hs.pos >= hs.len {
		return 0
	}
	ch = hs.s[hs.pos]
	if h5_is_white(ch) {
		hs.pos += 1
		return h5_state_before_attribute_name(hs)
	} else if ch == CHAR_SLASH {
		hs.pos += 1
		return h5_state_self_closing_start_tag(hs)
	} else if ch == CHAR_GT {
		hs.token_start = hs.pos
		hs.token_len = 1
		hs.token_type = TAG_NAME_CLOSE
		hs.pos += 1
		hs.state = h5_state_data
		return 1
	} else {
		return h5_state_before_attribute_name(hs)
	}
}

/**
 * 12.2.4.43
 */
func h5_state_self_closing_start_tag(hs *h5_state_t) int {
	var ch uint8

	if hs.pos >= hs.len {
		return 0
	}
	ch = hs.s[hs.pos]
	if ch == CHAR_GT {
		hs.token_start = hs.pos - 1
		hs.token_len = 2
		hs.token_type = TAG_NAME_SELFCLOSE
		hs.state = h5_state_data
		hs.pos += 1
		return 1
	} else {
		return h5_state_before_attribute_name(hs)
	}
}

/**
 * 12.2.4.44
 */
func h5_state_bogus_comment(hs *h5_state_t) int {
	idx := memchr(hs.s[hs.pos:], hs.pos, CHAR_GT)
	if idx == -1 {
		hs.token_start = hs.pos
		hs.token_len = hs.len - hs.pos
		hs.pos = hs.len
		hs.state = h5_state_eof
	} else {
		hs.token_start = hs.pos
		hs.token_len = (hs.len - idx) - hs.pos
		hs.pos = idx + 1
		hs.state = h5_state_data
	}

	hs.token_type = TAG_COMMENT
	return 1
}

/**
 * 12.2.4.44 ALT
 */
func h5_state_bogus_comment2(hs *h5_state_t) int {
	var pos int

	pos = hs.pos
	for {
		idx := memchr(hs.s[pos:], hs.pos, CHAR_PERCENT)
		if idx == -1 || idx+1 >= hs.len {
			hs.token_start = hs.pos
			hs.token_len = hs.len - hs.pos
			hs.pos = hs.len
			hs.token_type = TAG_COMMENT
			hs.state = h5_state_eof
			return 1
		}

		if hs.s[idx+1] != CHAR_GT {
			pos = idx + 1
			continue
		}

		/* ends in %> */
		hs.token_start = hs.pos
		hs.token_len = (hs.len - idx) - hs.pos
		hs.pos = idx + 2
		hs.state = h5_state_data
		hs.token_type = TAG_COMMENT
		return 1
	}
}

/**
 * 8.2.4.45
 */
func h5_state_markup_declaration_open(hs *h5_state_t) int {
	var remaining int
	remaining = hs.len - hs.pos
	if remaining >= 7 &&
		/* case insensitive */
		(hs.s[hs.pos+0] == 'D' || hs.s[hs.pos+0] == 'd') &&
		(hs.s[hs.pos+1] == 'O' || hs.s[hs.pos+1] == 'o') &&
		(hs.s[hs.pos+2] == 'C' || hs.s[hs.pos+2] == 'c') &&
		(hs.s[hs.pos+3] == 'T' || hs.s[hs.pos+3] == 't') &&
		(hs.s[hs.pos+4] == 'Y' || hs.s[hs.pos+4] == 'y') &&
		(hs.s[hs.pos+5] == 'P' || hs.s[hs.pos+5] == 'p') &&
		(hs.s[hs.pos+6] == 'E' || hs.s[hs.pos+6] == 'e') {
		return h5_state_doctype(hs)
	} else if remaining >= 7 &&
		/* upper case required */
		hs.s[hs.pos+0] == '[' &&
		hs.s[hs.pos+1] == 'C' &&
		hs.s[hs.pos+2] == 'D' &&
		hs.s[hs.pos+3] == 'A' &&
		hs.s[hs.pos+4] == 'T' &&
		hs.s[hs.pos+5] == 'A' &&
		hs.s[hs.pos+6] == '[' {
		hs.pos += 7
		return h5_state_cdata(hs)
	} else if remaining >= 2 &&
		hs.s[hs.pos+0] == '-' &&
		hs.s[hs.pos+1] == '-' {
		hs.pos += 2
		return h5_state_comment(hs)
	}

	return h5_state_bogus_comment(hs)
}

/**
 * 12.2.4.48
 * 12.2.4.49
 * 12.2.4.50
 * 12.2.4.51
 *   state machine spec is confusing since it can only look
 *   at one character at a time but simply it's comments end by:
 *   1) EOF
 *   2) ending in -.
 *   3) ending in -!>
 */
func h5_state_comment(hs *h5_state_t) int {
	var ch uint8
	var pos int
	var offset int
	end := hs.len

	pos = hs.pos
	for {

		idx := memchr(hs.s[pos:], hs.pos, CHAR_DASH)

		/* did not find anything or has less than 3 chars left */
		if idx == -1 || idx > hs.len-3 {
			hs.state = h5_state_eof
			hs.token_start = hs.pos
			hs.token_len = hs.len - hs.pos
			hs.token_type = TAG_COMMENT
			return 1
		}
		offset = 1
		/* skip all nulls */
		for idx+offset < end && hs.s[idx+offset] == 0 {
			offset += 1
		}
		if idx+offset == end {
			hs.state = h5_state_eof
			hs.token_start = hs.pos
			hs.token_len = hs.len - hs.pos
			hs.token_type = TAG_COMMENT
			return 1
		}

		ch = hs.s[idx+offset]
		if ch != CHAR_DASH && ch != CHAR_BANG {
			pos = idx + 1
			continue
		}
		offset += 1
		if idx+offset == end {
			hs.state = h5_state_eof
			hs.token_start = hs.pos
			hs.token_len = hs.len - hs.pos
			hs.token_type = TAG_COMMENT
			return 1
		}

		ch = hs.s[idx+offset]
		if ch != CHAR_GT {
			pos = idx + 1
			continue
		}
		offset += 1

		/* ends in -. or -!> */
		hs.token_start = hs.pos
		hs.token_len = (hs.len - idx) - hs.pos
		hs.pos = (idx + offset)
		hs.state = h5_state_data
		hs.token_type = TAG_COMMENT
		return 1
	}
}

func h5_state_cdata(hs *h5_state_t) int {
	var pos int
	pos = hs.pos
	for {
		idx := memchr(hs.s[pos:], hs.pos, CHAR_RIGHTB)

		/* did not find anything or has less than 3 chars left */
		if idx == -1 || idx > hs.len-3 {
			hs.state = h5_state_eof
			hs.token_start = hs.pos
			hs.token_len = hs.len - hs.pos
			hs.token_type = DATA_TEXT
			return 1
		} else if hs.s[idx+1] == CHAR_RIGHTB && hs.s[idx+2] == CHAR_GT {
			hs.state = h5_state_data
			hs.token_start = hs.pos
			hs.token_len = (hs.len - idx) - hs.pos
			hs.pos = idx + 3
			hs.token_type = DATA_TEXT
			return 1
		} else {
			pos = idx + 1
		}
	}
}

/**
 * 8.2.4.52
 * http://www.w3.org/html/wg/drafts/html/master/syntax.html#doctype-state
 */
func h5_state_doctype(hs *h5_state_t) int {
	hs.token_start = hs.pos
	hs.token_type = DOCTYPE

	idx := memchr(hs.s[hs.pos:], hs.pos, CHAR_GT)
	if idx == -1 {
		hs.state = h5_state_eof
		hs.token_len = hs.len - hs.pos
	} else {
		hs.state = h5_state_data
		hs.token_len = (hs.len - idx) - hs.pos
		hs.pos = idx + 1
	}
	return 1
}
