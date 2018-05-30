### grapeSQLI

grapeSQLI是一种简单易用的Sql inject & XSS分析程序。

兼容且使用libinjection指纹数据以及搜索模式。

libinjection已经拥有非常完美的思维模式，没必要颠覆它，所以我的大部分代码来自于libinjection，并针对GOLANG做出优化。


经过针对GO语言的优化，目前的性能在可以接受的范围内，具体请参考Benchmark节。


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
Benchmark_XSSParser-8   	 3000000	       458 ns/op	      80 B/op	       1 allocs/op
Benchmark_XSSParserParallel-8   	10000000	       150 ns/op	      80 B/op	       1 allocs/op
```

### SQLI例子
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