package GSQLI

type ptr_lookup_fn func(state *sqli_state, lookuptype int, word string, len int) uint8

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

	/*
	 * Pointer to token position in tokenvec, above
	 */
	current *sqli_token

	/*
	 * fingerprint pattern c-string
	 * +1 for ending null
	 * Minimum of 8 bytes to add gcc's -fstack-protector to work
	 */
	fingerprint [8]uint8

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

type sfilter = sqli_state

func SQLInject(src string) error {

	return nil
}
