package GSQLI

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

	xssWhites = []string{
		"123 LIKE -1234.5678E+2;",
		"APPLE 19.123 'FOO' \"BAR\"",
		"/* BAR */ UNION ALL SELECT (2,3,4)",
		"1 || COS(+0X04) --FOOBAR",
		"dog apple @cat banana bar",
		`dog apple cat \"banana \'bar`,
		"102 TABLE CLOTH",
		"(1001-'1') union select 1,2,3,4 from credit_cards",
	}
)

func Test_XSSParser(t *testing.T) {
	for _, tv := range xssToken {
		fmt.Println(tv)
		assert.Equal(t, XSSParser(tv), true)
	}
}

func Test_XSSParserWhite(t *testing.T) {
	for _, tv := range xssWhites {
		fmt.Println(tv)
		assert.Equal(t, XSSParser(tv), false)
	}
}

func Benchmark_XSSParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		XSSParser(xssToken[0])
	}
}

func Benchmark_XSSParserParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			XSSParser(xssToken[0])
		}
	})
}
