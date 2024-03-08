package main

import (
	"fmt"
	"encoding/json"
	"github.com/vjeantet/grok"
)

func main() {
	fmt.Println("# Default Capture :")
	g, _ := grok.New()
	values, _ := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}

	fmt.Println("\n# Named Capture :")
	g, _ = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	values, _ = g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}

	fmt.Println("\n# Add custom patterns :")
	// We add 3 patterns to our Grok instance, to structure an IRC message
	g, _ = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern("IRCUSER", `\A@(\w+)`)
	g.AddPattern("IRCBODY", `.*`)
	g.AddPattern("IRCMSG", `%{IRCUSER:user} .* : %{IRCBODY:message}`)
	values, _ = g.Parse("%{IRCMSG}", `@vjeantet said : Hello !`)
	for k, v := range values {
		fmt.Printf("%+15s: %s\n", k, v)
	}

	fmt.Println("\n# Parse into a Nested map")
	g, _ = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	nested_values,_ := g.ParseTyped("%{TIME:time_stamp}: %{USER:[name][first_name]} is %{POSINT:[person][age]:int} years old and %{NUMBER:[person][height]:float} meters tall",`12:23:31: bob is 23 years old and 4.2 meters tall`)

	j, _ := json.MarshalIndent(nested_values, "", "\t")
	fmt.Println(string(j))
}
