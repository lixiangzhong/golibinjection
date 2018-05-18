### grapeSQLI

grapeSQLI is easy to use Sql Inject & XSS Parser.For golang (like libinjection)

grapeSQLI use libinjection fingerprint data and search mode.

### don't use any cgo

cgo is bad idea,it's very slow and cgo is not go,So We Rewrite libinjection for golang.

### usage

```
    go get -u github.com/koangel/grapeSQLI
```

### xss example

```
package main

import (
    "github.com/koangel/grapeSQLI"
)

func main() {
    if GSQLI.XSSParser("<a href=\"  javascript:alert(1);\" >") {
        // todo something
    }
}

```

### xss benchmark

```
Benchmark_XSSParser-8   	 1000000	      1233 ns/op	     184 B/op	      27 allocs/op
Benchmark_XSSParserParallel-8   	 5000000	       349 ns/op	     184 B/op	      27 allocs/op
```