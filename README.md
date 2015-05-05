[![GoDoc](https://godoc.org/github.com/gemsi/grok?status.svg)](https://godoc.org/github.com/gemsi/grok)
[![Build Status](https://travis-ci.org/gemsi/grok.svg)](https://travis-ci.org/gemsi/grok)
[![Coverage Status](https://coveralls.io/repos/gemsi/grok/badge.png?branch=master)](https://coveralls.io/r/gemsi/grok?branch=master)
[![Documentation Status](https://readthedocs.org/projects/grok-lib-for-golang/badge/?version=latest)](https://readthedocs.org/projects/grok-lib-for-golang/?badge=latest)


# grok
A simple library to use/parse grok patterns in Go.

# Installation
Make sure you have a working Go environment.

```sh
go get github.com/gemsi/grok```

# Use in your project
```go
import "github.com/gemsi/grok"```

# Example
```go
package main

import (
  "fmt"
    "github.com/gemsi/grok"
)

func main() {
    g := grok.New()
  values, _ := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)

  for k, v := range values {
    fmt.Printf("%+15s: %s\n", k, v)
  }
}
```

output:
```
       response: 404
          bytes: 207
        request: /index.php
           auth: -
      timestamp: 23/Apr/2014:22:58:32 +0200
           verb: GET
    httpversion: 1.1
       clientip: 127.0.0.1
          ident: -
     rawrequest:
```
