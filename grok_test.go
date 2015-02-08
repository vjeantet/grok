package grok

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	defer g.Free()

	if g == nil && g.g == nil {
		t.Fatal("Failed to initialize grok")
	}
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
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to match!")
	}
	if &match.gm == nil {
		t.Fatal("Match object not correctly populated")
	}
}

func TestMatchCaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("./patterns/base")
	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"
	pattern := "%{DAY}"
	g.Compile(pattern)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}

	captures := match.Captures()
	if dayCap := captures["DAY"][0]; dayCap != "Tue" {
		t.Fatal("Day should equal Tue")
	}
}

func TestURICaptures(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPatternsFromFile("./patterns/base")
	text := "https://www.google.com/search?q=moose&sugexp=chrome,mod=16&sourceid=chrome&ie=UTF-8"
	pattern := "%{URI}"
	g.Compile(pattern)
	match := g.Match(text)
	if match == nil {
		t.Fatal("Unable to find match!")
	}

	captures := match.Captures()

	if host := captures["URIHOST"][0]; host != "www.google.com" {
		t.Fatal("URIHOST should be www.google.com")
	}
	if path := captures["URIPATH"][0]; path != "/search" {
		t.Fatal("URIPATH should be /search")
	}
}

func TestDiscovery(t *testing.T) {
	g := New()
	defer g.Free()

	g.AddPattern("IP", "(?<![0-9])(?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})[.](?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?![0-9])")

	text := "1.2.3.4"
	discovery := g.Discover(text)
	g.Compile(discovery)
	captures := g.Match(text).Captures()
	if ip := captures["IP"][0]; ip != text {
		t.Fatal("IP should be 1.2.3.4")
	}
}

func TestPileMatching(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPattern("foo", ".*(foo).*")
	p.AddPattern("bar", ".*(bar).*")

	p.Compile("%{bar}")

	grok, match := p.Match("bar")

	captures := match.Captures()
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern")
	}

	captures = grok.Match("bar").Captures()
	if bar := captures["bar"][0]; bar != "bar" {
		t.Fatal("Should match the bar pattern")
	}
}

func TestPileAddPatternsFromFile(t *testing.T) {
	p := NewPile()
	defer p.Free()

	p.AddPatternsFromFile("./patterns/base")
	p.Compile("%{DAY}")

	text := "Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157"

	_, match := p.Match(text)

	captures := match.Captures()
	if day := captures["DAY"][0]; day != "Tue" {
		t.Fatal("Should match the Tue")
	}
}
