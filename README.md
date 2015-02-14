[![GoDoc](https://godoc.org/github.com/gemsi/grok?status.svg)](https://godoc.org/github.com/gemsi/grok)
[![Build Status](https://travis-ci.org/gemsi/grok.svg)](https://travis-ci.org/gemsi/grok)
[![Coverage Status](https://coveralls.io/repos/gemsi/grok/badge.png?branch=master)](https://coveralls.io/r/gemsi/grok?branch=master)

# grok
simple library to use/parse grok patterns with go

# Installation
Make sure you have the a working Go environment.
```go get github.com/gemsi/grok```

# Use in your project
```import "github.com/gemsi/grok"```


# Exemple
```
package main

import (
	"fmt"

	"github.com/gemsi/grok"
)

func main() {
	g := grok.New()
	g.AddPatternsFromPath("../patterns") // path to pattern folder
	values, _ := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)

	for k, v := range values {
		fmt.Printf("%+12s: %s\n", k, v)
	}
}
```

Will print
```
    HOSTNAME: 127.0.0.1
          IP: 
        IPV6: 
   timestamp: 23/Apr/2014:22:58:32 +0200
       MONTH: Apr
        HOUR: 22
         INT: +0200
COMMONAPACHELOG: 127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207
 httpversion: 1.1
   BASE10NUM: 207
     request: /index.php
        YEAR: 2014
        verb: GET
    response: 404
       ident: -
        TIME: 22:58:32
      MINUTE: 58
    MONTHDAY: 23
    clientip: 127.0.0.1
        IPV4: 
    USERNAME: -
        auth: -
      SECOND: 32
  rawrequest: 
       bytes: 207
            : 207
```

# TODO :
* Use default patterns if AddPatternsFromPath is not used 