[![GoDoc](https://godoc.org/github.com/gemsi/grok?status.svg)](https://godoc.org/github.com/gemsi/grok)
[![Build Status](https://travis-ci.org/gemsi/grok.svg)](https://travis-ci.org/gemsi/grok)
[![Coverage Status](https://coveralls.io/repos/gemsi/grok/badge.png?branch=master)](https://coveralls.io/r/gemsi/grok?branch=master)
[![Documentation Status](https://readthedocs.org/projects/grok-lib-for-golang/badge/?version=latest)](https://readthedocs.org/projects/grok-lib-for-golang/?badge=latest)
         

# grok
simple library to use/parse grok patterns with go

# Installation
Make sure you have the a working Go environment.

```go get github.com/gemsi/grok```

# Use in your project
```import "github.com/gemsi/grok"```


# Example
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
		fmt.Printf("%+15s: %s\n", k, v)
	}
}
```

output:
```
       response: 404
          bytes: 207
               : 207
       HOSTNAME: 127.0.0.1
       USERNAME: -
       MONTHDAY: 23
        request: /index.php
      BASE10NUM: 207
           IPV6: 
           auth: -
      timestamp: 23/Apr/2014:22:58:32 +0200
           verb: GET
    httpversion: 1.1
           TIME: 22:58:32
           HOUR: 22
COMMONAPACHELOG: 127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207
       clientip: 127.0.0.1
             IP: 
          ident: -
          MONTH: Apr
           YEAR: 2014
         SECOND: 32
            INT: +0200
           IPV4: 
         MINUTE: 58
     rawrequest: 
```

# TODO
* Use default patterns if AddPatternsFromPath is not used 
