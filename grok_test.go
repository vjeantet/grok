package grok

import "testing"

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("error")
	}

}

func TestAddPattern(t *testing.T) {
	g := New()

	name := "DAY"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"

	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)

	if len(g.Patterns()) != 2 {
		t.Fatal("two pattern should be available")
	}
}

func TestDayCompile(t *testing.T) {
	g := New()

	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	//	g.AddPattern("MONTH", "Aout")

	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestDayCompileAndMatch(t *testing.T) {
	g := New()

	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"

	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}

	if !g.Match(text) {
		t.Fatal("text does not match pattern ! %s", err)
	}
}

func TestCaptures(t *testing.T) {
	g := New()

	//g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	g.AddPatternsFromFile("./patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	g.Compile(pattern)

	if !g.Match(text) {
		t.Fatal("text does not match pattern !")
	}

	captures, err := g.Captures(text)
	if err != nil {
		t.Fatal("Unable to capture ! %s", err)
	}
	if dayCap := captures["DAY"]; dayCap != "Tue" {
		t.Fatal("Day should equal Tue", err)
	}
}

func TestAddPatternsFromFile(t *testing.T) {
	//g := New()
	//g.AddPatternsFromFile("./patterns/base")
}

// func TestDateCaptures(t *testing.T) {
// 	g := New()

// 	g.AddPatternsFromFile("./patterns/base")
// 	text := "2014/12/27"

// 	pattern := "%{DATE_EU}"
// 	g.Compile(pattern)

// 	if !g.Match(text) {
// 		t.Fatal("text does not match pattern!")
// 	}

// 	captures, err := g.Captures(text)
// 	if err != nil {
// 		t.Fatal("text does not match pattern!", err)
// 	}
// 	log.Printf("%s", captures)
// }

// func TestDiscovery(t *testing.T) {
// 	g := New()
// 	defer g.Free()

// 	g.AddPattern("IP", "(?<![0-9])(?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?![0-9])")

// 	text := "1.2.3.4"
// 	discovery, err := g.Discover(text)
// 	if err != nil {
// 		t.Fatal("Unable to dicover !", err)
// 	}
// 	g.Compile(discovery)
// 	match, err := g.Match(text)
// 	if err != nil {
// 		t.Fatal("match error !", err)
// 	}

// 	captures, err := match.Captures()
// 	if err != nil {
// 		t.Fatal("Unable to capture!", err)
// 	}
// 	if ip := captures["IP"][0]; ip != text {
// 		t.Fatal("IP should be 1.2.3.4", err)
// 	}
// }

// func TestPileMatching(t *testing.T) {
// 	p := NewPile()
// 	defer p.Free()

// 	p.AddPattern("foo", ".*(foo).*")
// 	p.AddPattern("bar", ".*(bar).*")

// 	p.Compile("%{bar}")

// 	grok, match := p.Match("bar")

// 	captures, err := match.Captures()
// 	if err != nil {
// 		t.Fatal("Unable to capture!", err)
// 	}
// 	if bar := captures["bar"][0]; bar != "bar" {
// 		t.Fatal("Should match the bar pattern", err)
// 	}

// 	match2, err := grok.Match("bar")
// 	if err != nil {
// 		t.Fatal("match error !", err)
// 	}

// 	captures, err = match2.Captures()
// 	if err != nil {
// 		t.Fatal("Unable to capture!", err)
// 	}
// 	if bar := captures["bar"][0]; bar != "bar" {
// 		t.Fatal("Should match the bar pattern", err)
// 	}
// }

// func TestPileAddPatternsFromFile(t *testing.T) {
// 	p := NewPile()
// 	defer p.Free()

// 	p.AddPatternsFromFile("./patterns/base")
// 	p.Compile("%{DAY}")

// 	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"

// 	_, match := p.Match(text)

// 	captures, err := match.Captures()
// 	if err != nil {
// 		t.Fatal("Unable to capture!", err)
// 	}
// 	if day := captures["DAY"][0]; day != "Tue" {
// 		t.Fatal("Should match the Tue", err)
// 	}
// }
