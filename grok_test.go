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

	if len(g.patterns) != 2 {
		t.Fatal("two pattern should be available")
	}
}

func TestDayCompile(t *testing.T) {
	g := New()
	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestAddPatternsFromFile(t *testing.T) {
	g := New()
	g.AddPatternsFromFile("./patterns/base")
}

func BenchmarkCaptures(t *testing.B) {
	g := New()
	g.AddPatternsFromFile("./patterns/base")

	check := func(key, value, pattern, text string) {
		if err := g.Compile(pattern); err != nil {
			t.Fatalf("error can not compile : %s", err.Error())
		}

		if captures, err := g.Captures(text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	// run the check function b.N times
	for n := 0; n < t.N; n++ {
		check("verb", "GET",
			"%{COMMONAPACHELOG}",
			`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
		)
	}
}

func TestNamedCaptures(t *testing.T) {
	g := New()
	g.AddPatternsFromFile("./patterns/base")

	check := func(key, value, pattern, text string) {
		g.Compile(pattern)
		captures, _ := g.Captures(text)
		if captures[key] != value {
			t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
		}
	}

	check("jour", "Tue",
		"%{DAY:jour}",
		"Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157",
	)
}

func TestCaptures(t *testing.T) {
	g := New()
	g.AddPatternsFromFile("./patterns/base")

	check := func(key, value, pattern, text string) {
		if err := g.Compile(pattern); err != nil {
			t.Fatalf("error can not compile : %s", err.Error())
		}

		if captures, err := g.Captures(text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	check("DAY", "Tue",
		"%{DAY}",
		"Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157",
	)
	check("jour", "Tue",
		"%{DAY:jour}",
		"Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157",
	)
	check("clientip", "127.0.0.1",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("verb", "GET",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("TIME", "22:58:32",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("timestamp", "23/Apr/2014:22:58:32 +0200",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("bytes", "207",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)

}
