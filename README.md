### grapeSQLI

grapeSQLI is easy to use Sql Inject & XSS Parser.For golang (like libinjection)

grapeSQLI use libinjection fingerprint data and search mode.

### don't use any cgo

cgo is bad idea,it's very slow and cgo is not go,So We Rewrite libinjection for golang.

### usage

```
    go get -u github.com/koangel/grapeSQLI
```

### example

```
package main

import (
    "github.com/koangel/grapeSQLI"
)
```