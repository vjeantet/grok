package main

import (
	"fmt"

	"github.com/gemsi/grok"
)

func main() {
	fmt.Println("# Default Capture :\n")
	g := grok.New()
	values, _ := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}

	fmt.Println("\n\n# Named Capture :\n")
	g = grok.New(grok.NAMEDCAPTURE)
	values, _ = g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}

	fmt.Println("\n\n# Add custom patterns :\n")
	// We add 3 patterns to our Grok instance, to structure an IRC message
	g = grok.New(grok.NAMEDCAPTURE)
	g.AddPattern("IRCUSER", `\A@(\w+)`)
	g.AddPattern("IRCBODY", `.*`)
	g.AddPattern("IRCMSG", `%{IRCUSER:user} .* : %{IRCBODY:message}`)
	values, _ = g.Parse("%{IRCMSG}", `@vjeantet said : Hello !`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}
}
