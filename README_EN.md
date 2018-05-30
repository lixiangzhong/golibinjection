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
Benchmark_XSSParser-8   	 3000000	       458 ns/op	      80 B/op	       1 allocs/op
Benchmark_XSSParserParallel-8   	10000000	       150 ns/op	      80 B/op	       1 allocs/op
```

### SQLI example
```
package main

import (
    "github.com/koangel/grapeSQLI"
)

func main() {
    if GSQLI.SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--") {
        // todo something
    }
}
```

### SQLI Benchmark

```
BenchmarkSQLInject-8   	  300000	      5019 ns/op	    1376 B/op	      61 allocs/op
BenchmarkSQLInjectParallel-8   	 1000000	      2873 ns/op	    1376 B/op	      61 allocs/op
```