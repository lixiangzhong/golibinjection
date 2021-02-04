```
                             ____   ___  _     ____       
   __ _ _ __ __ _ _ __   ___/ ___| / _ \| |   / ___| ___  
  / _` | '__/ _` | '_ \ / _ \___ \| | | | |  | |  _ / _ \ 
 | (_| | | | (_| | |_) |  __/___) | |_| | |__| |_| | (_) |
  \__, |_|  \__,_| .__/ \___|____/ \__\_\_____\____|\___/   
  |___/          |_|                     grapeSQLI is easy to use Sql Inject & XSS Parser
```
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![Open Source Love svg2](https://badges.frapsoft.com/os/v2/open-source.svg?v=103)](https://github.com/ellerbrock/open-source-badges/)

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
    if err:= GSQLI.SQLInject("asdf asd ; -1' and 1=1 union/* foo */select load_file('/etc/passwd')--");err != nil {
        // todo something
    }
}
```

### SQLI Benchmark

```
BenchmarkSQLInject-8   	  300000	      5019 ns/op	    1376 B/op	      61 allocs/op
BenchmarkSQLInjectParallel-8   	 1000000	      2873 ns/op	    1376 B/op	      61 allocs/op
```


## **Thanks**

Use Jetbrains Ide for project

[![saythanks](https://img.shields.io/badge/say-thanks-ff69b4.svg)](https://saythanks.io/to/kennethreitz)
[![Generic badge](https://img.shields.io/badge/JetBrains-Goland-<COLOR>.svg)](https://shields.io/)
[![Generic badge](https://img.shields.io/badge/JetBrains-CLion-<COLOR>.svg)](https://shields.io/)


