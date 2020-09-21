package main

import (
	"net/http"
	_ "net/http/pprof"

	gsqli "github.com/koangel/grapeSQLI"
)

var (
	xssToken = []string{
		"<a href=\"  javascript:alert(1);\" >",
		"<a href=\"JAVASCRIPT:alert(1);\" >",
		"<a href=javascript:alert(1)>",
		"<a href=\"javascript:alert(1)\">",
		"<a href='javascript:alert(1)'>",
		"<a href  =   javascript:alert(1); >",
		"red;</style><script>alert(1);</script>",
		"red;}</style><script>alert(1);</script>",
		"red;\"/><script>alert(1);</script>",
		"<script>alert(1);</script>",
		"><script>alert(1);</script>",
		"x ><script>alert(1);</script>",
		"' ><script>alert(1);</script>",
		"\"><script>alert(1);</script>",
		"');}</style><script>alert(1);</script>",
		"onerror=alert(1)>",
		"x onerror=alert(1);>",
		"x' onerror=alert(1);>",
		"x\" onerror=alert(1);>",
	}
)

func main() {
	go func() {
		http.ListenAndServe(":6602", nil)
	}()

	// test XSS
	/*for {
		for _, tv := range xssToken {
			GSQLI.XSSParser(tv)
		}
	}*/

	for {
		gsqli.SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
	}
}
