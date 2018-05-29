### grapeSQLI

grapeSQLI is easy to use Sql Inject & XSS Parser.For pure go (like libinjection)

grapeSQLI use libinjection fingerprint data and search mode.

libinjection is great and perfect library,so we just translate related code to golang. but we optimize the code for golang.

grapeSQLI is pure go library.


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