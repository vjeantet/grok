package grok

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	defer g.Free()

}

func TestDayCompile(t *testing.T) {
	g := New()
	defer g.Free()

	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestDayCompileAndMatch(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("./patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
	_, err = g.Match(text)
	if err != nil {
		t.Fatal("Unable to match!", err)
	}
}

func TestMatchCaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("./patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	g.Compile(pattern)
	match, err := g.Match(text)
	if err != nil {
		t.Fatal("Unable to find match!", err)
	}

	captures, err := match.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if dayCap := captures["DAY"][0]; dayCap != "Tue" {
		t.Fatal("Day should equal Tue", err)
	}
}

func TestURICaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("./patterns/base")
	text := "https://www.google.com/search?q=moose&sugexp=chrome,mod=16&sourceid=chrome&ie=UTF-8"
	pattern := "%{URI}"
	g.Compile(pattern)
	match, err := g.Match(text)
	if err != nil {
		t.Fatal("Unable to find match!", err)
	}

	captures, err := match.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if host := captures["URIHOST"][0]; host != "www.google.com" {
		t.Fatal("URIHOST should be www.google.com", err)
	}
	if path := captures["URIPATH"][0]; path != "/search" {
		t.Fatal("URIPATH should be /search", err)
	}
}

func TestDiscovery(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPattern("IP", "(?<![0-9])(?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?![0-9])")

	text := "1.2.3.4"
	discovery, err := g.Discover(text)
	if err != nil {
		t.Fatal("Unable to dicover !", err)
	}
	g.Compile(discovery)
	match, err := g.Match(text)
	if err != nil {
		t.Fatal("match error !", err)
	}

	captures, err := match.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if ip := captures["IP"][0]; ip != text {
		t.Fatal("IP should be 1.2.3.4", err)
	}
}

func TestPileMatching(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPattern("foo", ".*(foo).*")
	p.AddPattern("bar", ".*(bar).*")

	p.Compile("%{bar}")

	grok, match := p.Match("bar")

	captures, err := match.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern", err)
	}

	match2, err := grok.Match("bar")
	if err != nil {
		t.Fatal("match error !", err)
	}

	captures, err = match2.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern", err)
	}
}

func TestPileAddPatternsFromFile(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPatternsFromFile("./patterns/base")
	p.Compile("%{DAY}")

	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"

	_, match := p.Match(text)

	captures, err := match.Captures()
	if err != nil {
		t.Fatal("Unable to capture!", err)
	}
	if day := captures["DAY"][0]; day != "Tue" {
		t.Fatal("Should match the Tue", err)
	}
}
