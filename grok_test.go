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
	name := "DAYO"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"
	c_patterns := len(g.patterns)
	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)

	if len(g.patterns) != c_patterns+2 {
		t.Fatalf("%d patterns should be available, have %d", c_patterns+2, len(g.patterns))
	}
}

func TestMatch(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")

	if r, err := g.Match("%{MONTH}", "June"); !r {
		t.Fatalf("June should match %s: err=%s", "%{MONTH}", err.Error())
	}

}
func TestDoesNotMatch(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	if r, _ := g.Match("%{MONTH}", "13"); r {
		t.Fatalf("13 should not match %s", "%{MONTH}")
	}
}

func TestErrorMatch(t *testing.T) {
	g := New()
	if _, err := g.Match("(", "13"); err == nil {
		t.Fatal("Error expected")
	}

}

func TestDayCompile(t *testing.T) {
	g := New()
	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	pattern := "%{DAY}"
	_, err := g.compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestErrorCompile(t *testing.T) {
	g := New()
	_, err := g.compile("(")
	if err == nil {
		t.Fatal("Error:", err)
	}
}

func BenchmarkCaptures(t *testing.B) {
	g := New()
	g.AddPatternsFromPath("./patterns/base")

	check := func(key, value, pattern, text string) {

		if captures, err := g.Parse(pattern, text); err != nil {
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
	g.AddPatternsFromPath("./patterns")

	check := func(key, value, pattern, text string) {
		captures, _ := g.Parse(pattern, text)
		if captures[key] != value {
			t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
		}
	}

	check("jour", "Tue",
		"%{DAY:jour}",
		"Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157",
	)
}

func TestErrorCaptureUnknowPattern(t *testing.T) {
	g := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.Parse(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParse(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	res, _ := g.Parse("%{DAY:day}", "Tue qds")
	if res["day"] != "Tue" {
		t.Fatalf("DAY should be 'Tue' have '%s'", res["day"])
	}
}

func TestErrorParseToMultiMap(t *testing.T) {
	g := New()
	pattern := "%{UNKNOWNPATTERN:up}"
	_, err := g.ParseToMultiMap(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParseToMultiMap(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	res, _ := g.ParseToMultiMap("%{DAY:day} %{DAY:day} %{DAY:day}", "Tue Wed Fri")
	if len(res["day"]) != 3 {
		t.Fatalf("day[] should be an array of 3 elements, but is '%s'", res["day"])
	}
	if res["day"][0] != "Tue" {
		t.Fatalf("day[0] should be 'Tue' have '%s'", res["DAY"][0])
	}
	if res["day"][1] != "Wed" {
		t.Fatalf("day[1] should be 'Wed' have '%s'", res["DAY"][1])
	}
	if res["day"][2] != "Fri" {
		t.Fatalf("day[2] should be 'Fri' have '%s'", res["DAY"][2])
	}
}

func TestCaptures(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")

	check := func(key, value, pattern, text string) {

		if captures, err := g.Parse(pattern, text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	check("day", "Tue",
		"%{DAY:day}",
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
	check("timestamp", "23/Apr/2014:22:58:32 +0200",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("bytes", "207",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)

	//PATH
	check("path", `c:\winfows\sdf.txt`, "%{WINPATH:path}", `s dfqs c:\winfows\sdf.txt`)
	check("path", `\\sdf\winfows\sdf.txt`, "%{WINPATH:path}", `s dfqs \\sdf\winfows\sdf.txt`)
	check("path", `/usr/lib/`, "%{UNIXPATH:path}", `s dfqs /usr/lib/ sqfd`)
	check("path", `/usr/lib`, "%{UNIXPATH:path}", `s dfqs /usr/lib sqfd`)
	check("path", `/usr/`, "%{UNIXPATH:path}", `s dfqs /usr/ sqfd`)
	check("path", `/usr`, "%{UNIXPATH:path}", `s dfqs /usr sqfd`)
	check("path", `/`, "%{UNIXPATH:path}", `s dfqs / sqfd`)

	//YEAR
	check("year", `4999`, "%{YEAR:year}", `s d9fq4999s ../ sdf`)
	check("year", `79`, "%{YEAR:year}", `s d79fq4999s ../ sdf`)
	check("timestamp", `2013-11-06 04:50:17,1599`, "%{TIMESTAMP_ISO8601:timestamp}", `s d9fq4999s ../ sdf 2013-11-06 04:50:17,1599sd`)

	//MAC
	check("mac", `01:02:03:04:ab:cf`, "%{MAC:mac}", `s d9fq4999s ../ sdf 2013- 01:02:03:04:ab:cf  11-06 04:50:17,1599sd`)
	check("mac", `01-02-03-04-ab-cd`, "%{MAC:mac}", `s d9fq4999s ../ sdf 2013- 01-02-03-04-ab-cd  11-06 04:50:17,1599sd`)

	//QUOTEDSTRING
	check("qs", `"lkj"`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd fk"lkj"mkj`)
	check("qs", `'lkj'`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd fk'lkj'mkj`)
	check("qs", `"fk'lkj'm"`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd "fk'lkj'm"kj`)
	check("qs", `'fk"lkj"m'`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd 'fk"lkj"m'kj`)
}

// Should be run with -race
func TestConcurentParse(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")

	check := func(key, value, pattern, text string) {

		if captures, err := g.Parse(pattern, text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	go check("qs", `"lkj"`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd fk"lkj"mkj`)
	go check("qs", `'lkj'`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd fk'lkj'mkj`)
	go check("qs", `'lkj'`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd fk'lkj'mkj`)
	go check("qs", `"fk'lkj'm"`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd "fk'lkj'm"kj`)
	go check("qs", `'fk"lkj"m'`, "%{QUOTEDSTRING:qs}", `qsdklfjqsd 'fk"lkj"m'kj`)
}
