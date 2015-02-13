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

func TestMatch(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	g.Compile("%{MONTH}")
	if r, _ := g.Match("June"); !r {
		t.Fatal("June should match %{MONTH}")
	}

}
func TestDoesNotMatch(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	g.Compile("%{MONTH}")
	if r, _ := g.Match("13"); r {
		t.Fatal("13 should not match %{MONTH}")
	}
}
func TestErrorMatchWithoutCompilation(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	if _, err := g.Match("12"); err == nil {
		t.Fatal("Match should return an error")
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

func BenchmarkCaptures(t *testing.B) {
	g := New()
	g.AddPatternsFromPath("./patterns/base")

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
	g.AddPatternsFromPath("./patterns")

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

func TestErrorCaptureUnknowPattern(t *testing.T) {
	g := New()
	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	pattern := "%{MONTH}"
	err := g.Compile(pattern)
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestErrorCompileRegex(t *testing.T) {
	g := New()
	g.AddPattern("DAY", "(")
	pattern := "%{DAY}"
	err := g.Compile(pattern)
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParse(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	res, _ := g.Parse("%{DAY}", "Tue qds")
	if res["DAY"] != "Tue" {
		t.Fatalf("DAY should be 'Tue' have '%s'", res["DAY"])
	}
}

func TestCapturesWithoutCompilation(t *testing.T) {
	g := New()
	g.AddPattern("DAY", "(")
	_, err := g.Captures("Lorem ipsum Ut officia ut minim ea laborum Excepteur ut consequat et pariatur nostrud.")

	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestCaptures(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")

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

	//PATH
	check("WINPATH", `c:\winfows\sdf.txt`, "%{WINPATH}", `s dfqs c:\winfows\sdf.txt`)
	check("WINPATH", `\\sdf\winfows\sdf.txt`, "%{WINPATH}", `s dfqs \\sdf\winfows\sdf.txt`)
	check("UNIXPATH", `/usr/lib/`, "%{UNIXPATH}", `s dfqs /usr/lib/ sqfd`)
	check("UNIXPATH", `/usr/lib`, "%{UNIXPATH}", `s dfqs /usr/lib sqfd`)
	check("UNIXPATH", `/usr/`, "%{UNIXPATH}", `s dfqs /usr/ sqfd`)
	check("UNIXPATH", `/usr`, "%{UNIXPATH}", `s dfqs /usr sqfd`)
	check("UNIXPATH", `/`, "%{UNIXPATH}", `s dfqs / sqfd`)

	//YEAR
	check("YEAR", `4999`, "%{YEAR}", `s d9fq4999s ../ sdf`)
	check("YEAR", `79`, "%{YEAR}", `s d79fq4999s ../ sdf`)
	check("TIMESTAMP_ISO8601", `2013-11-06 04:50:17,1599`, "%{TIMESTAMP_ISO8601}", `s d9fq4999s ../ sdf 2013-11-06 04:50:17,1599sd`)
	check("SECOND", `17,1599`, "%{TIMESTAMP_ISO8601}", `s d9fq4999s ../ sdf 2013-11-06 04:50:17,1599sd`)

	//MAC
	check("MAC", `01:02:03:04:ab:cf`, "%{MAC}", `s d9fq4999s ../ sdf 2013- 01:02:03:04:ab:cf  11-06 04:50:17,1599sd`)
	check("MAC", `01-02-03-04-ab-cd`, "%{MAC}", `s d9fq4999s ../ sdf 2013- 01-02-03-04-ab-cd  11-06 04:50:17,1599sd`)

	//QUOTEDSTRING
	check("QUOTEDSTRING", `"lkj"`, "%{QUOTEDSTRING}", `qsdklfjqsd fk"lkj"mkj`)
	check("QUOTEDSTRING", `'lkj'`, "%{QUOTEDSTRING}", `qsdklfjqsd fk'lkj'mkj`)
	check("QUOTEDSTRING", `"fk'lkj'm"`, "%{QUOTEDSTRING}", `qsdklfjqsd "fk'lkj'm"kj`)
	check("QUOTEDSTRING", `'fk"lkj"m'`, "%{QUOTEDSTRING}", `qsdklfjqsd 'fk"lkj"m'kj`)

}
