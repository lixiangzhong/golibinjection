### grapeSQLI

grapeSQLI是一种简单易用的Sql inject & XSS分析程序。

兼容且使用libinjection指纹数据以及搜索模式。

### 不要使用任何CGO

CGO是一个特别糟糕的东西，并且他很慢也不是GO的核心内容，他是一种社区兼容的妥协。
所以我们为GOLANG重写了libinjection。

### 用法

```
    go get -u github.com/koangel/grapeSQLI
```


### xss例子

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