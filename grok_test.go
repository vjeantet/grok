package grok

import "testing"

func TestNew(t *testing.T) {
	g := New()
	if len(g.Patterns()) == 0 {
		t.Fatal("the Grok object should have some patterns pre loaded")
	}

	g = NewWithConfig(&Config{NamedCapturesOnly: true})
	if len(g.Patterns()) == 0 {
		t.Fatal("the Grok object should have some patterns pre loaded")
	}
}

func TestParseWithDefaultCaptureMode(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["timestamp"] != "23/Apr/2014:22:58:32 +0200" {
			t.Fatalf("%s should be '%s' have '%s'", "timestamp", "23/Apr/2014:22:58:32 +0200", captures["timestamp"])
		}
		if captures["TIME"] != "" {
			t.Fatalf("%s should be '%s' have '%s'", "TIME", "", captures["TIME"])
		}
	}

	g = New()
	if captures, err := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["timestamp"] != "23/Apr/2014:22:58:32 +0200" {
			t.Fatalf("%s should be '%s' have '%s'", "timestamp", "23/Apr/2014:22:58:32 +0200", captures["timestamp"])
		}
		if captures["TIME"] != "22:58:32" {
			t.Fatalf("%s should be '%s' have '%s'", "TIME", "22:58:32", captures["TIME"])
		}
	}
}

func TestMultiParseWithDefaultCaptureMode(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	res, _ := g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:23:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["TIME"]) != 0 {
		t.Fatalf("DAY should be an array of 0 elements, but is '%s'", res["TIME"])
	}

	g = New()
	res, _ = g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:23:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["TIME"]) != 2 {
		t.Fatalf("TIME should be an array of 2 elements, but is '%s'", res["TIME"])
	}
	if len(res["timestamp"]) != 2 {
		t.Fatalf("timestamp should be an array of 2 elements, but is '%s'", res["timestamp"])
	}
}

func TestNewWithNoDefaultPatterns(t *testing.T) {
	g := NewWithConfig(&Config{SkipDefaultPatterns: true})
	if len(g.Patterns()) != 0 {
		t.Fatal("Using SkipDefaultPatterns the Grok object should not have any patterns pre loaded")
	}
}

func TestAddPatternsFromPath(t *testing.T) {
	g := New()
	err := g.AddPatternsFromPath("./Lorem ipsum Minim qui in.")
	if err == nil {
		t.Fatalf("AddPatternsFromPath should returns an error when path is invalid")
	}
}

func TestAddPatternsFromPathFileOpenErr(t *testing.T) {
	t.Skipped()
}

func TestAddPatternsFromPathFile(t *testing.T) {
	g := New()
	err := g.AddPatternsFromPath("./patterns/base")
	if err != nil {
		t.Fatalf("err %#v", err)
	}
}

func TestAddPatternErr(t *testing.T) {
	name := "Error"
	pattern := "%{ERR}"

	g := New()
	err := g.AddPattern(name, pattern)
	if err == nil {
		t.Fatalf("AddPattern should returns an error when path is invalid")
	}
}

func TestAddPattern(t *testing.T) {
	name := "DAYO"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"

	g := New()
	cPatterns := len(g.patterns)
	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)
	if len(g.patterns) != cPatterns+2 {
		t.Fatalf("%d Default patterns should be available, have %d", cPatterns+2, len(g.patterns))
	}

	g = NewWithConfig(&Config{NamedCapturesOnly: true})
	cPatterns = len(g.patterns)
	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)
	if len(g.patterns) != cPatterns+2 {
		t.Fatalf("%d NamedCapture patterns should be available, have %d", cPatterns+2, len(g.patterns))
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
	res, _ := g.Parse("%{DAY}", "Tue qds")
	if res["DAY"] != "Tue" {
		t.Fatalf("DAY should be 'Tue' have '%s'", res["DAY"])
	}
}

func TestErrorParseToMultiMap(t *testing.T) {
	g := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.ParseToMultiMap(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParseToMultiMap(t *testing.T) {
	g := New()
	g.AddPatternsFromPath("./patterns")
	res, _ := g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:23:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["TIME"]) != 2 {
		t.Fatalf("DAY should be an array of 3 elements, but is '%s'", res["TIME"])
	}
	if res["TIME"][0] != "23:58:32" {
		t.Fatalf("TIME[0] should be '23:58:32' have '%s'", res["TIME"][0])
	}
	if res["TIME"][1] != "22:58:32" {
		t.Fatalf("TIME[1] should be '22:58:32' have '%s'", res["TIME"][1])
	}
}

func TestParseToMultiMapOnlyNamedCaptures(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	g.AddPatternsFromPath("./patterns")
	res, _ := g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["timestamp"]) != 2 {
		t.Fatalf("timestamp should be an array of 2 elements, but is '%s'", res["timestamp"])
	}
	if res["timestamp"][0] != "23/Apr/2014:22:58:32 +0200" {
		t.Fatalf("timestamp[0] should be '23/Apr/2014:22:58:32 +0200' have '%s'", res["DAY"][0])
	}
	if res["timestamp"][1] != "24/Apr/2014:22:58:32 +0200" {
		t.Fatalf("timestamp[1] should be '24/Apr/2014:22:58:32 +0200' have '%s'", res["DAY"][1])
	}
}

func TestCaptureAll(t *testing.T) {
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

	check("timestamp", "23/Apr/2014:22:58:32 +0200",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("TIME", "22:58:32",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("SECOND", `17,1599`, "%{TIMESTAMP_ISO8601}", `s d9fq4999s ../ sdf 2013-11-06 04:50:17,1599sd`)
	check("HOSTNAME", `google.com`, "%{HOSTPORT}", `google.com:8080`)
	//HOSTPORT
	check("POSINT", `8080`, "%{HOSTPORT}", `google.com:8080`)
}

func TestNamedCapture(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
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

	check("timestamp", "23/Apr/2014:22:58:32 +0200",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("TIME", "",
		"%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)
	check("SECOND", ``, "%{TIMESTAMP_ISO8601}", `s d9fq4999s ../ sdf 2013-11-06 04:50:17,1599sd`)
	check("HOSTNAME", ``, "%{HOSTPORT}", `google.com:8080`)
	//HOSTPORT
	check("POSINT", ``, "%{HOSTPORT}", `google.com:8080`)
}

func TestCapturesAndNamedCapture(t *testing.T) {

	check := func(key, value, pattern, text string) {
		g := New()
		if captures, err := g.Parse(pattern, text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	checkNamed := func(key, value, pattern, text string) {
		g := NewWithConfig(&Config{NamedCapturesOnly: true})
		if captures, err := g.Parse(pattern, text); err != nil {
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
	checkNamed("jour", "Tue",
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

	//MAC
	check("MAC", `01:02:03:04:ab:cf`, "%{MAC}", `s d9fq4999s ../ sdf 2013- 01:02:03:04:ab:cf  11-06 04:50:17,1599sd`)
	check("MAC", `01-02-03-04-ab-cd`, "%{MAC}", `s d9fq4999s ../ sdf 2013- 01-02-03-04-ab-cd  11-06 04:50:17,1599sd`)

	//QUOTEDSTRING
	check("QUOTEDSTRING", `"lkj"`, "%{QUOTEDSTRING}", `qsdklfjqsd fk"lkj"mkj`)
	check("QUOTEDSTRING", `'lkj'`, "%{QUOTEDSTRING}", `qsdklfjqsd fk'lkj'mkj`)
	check("QUOTEDSTRING", `"fk'lkj'm"`, "%{QUOTEDSTRING}", `qsdklfjqsd "fk'lkj'm"kj`)
	check("QUOTEDSTRING", `'fk"lkj"m'`, "%{QUOTEDSTRING}", `qsdklfjqsd 'fk"lkj"m'kj`)

	//BASE10NUM
	check("BASE10NUM", `1`, "%{BASE10NUM}", `1`) // this is a nice one
	check("BASE10NUM", `8080`, "%{BASE10NUM}", `qsfd8080qsfd`)

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

	go check("QUOTEDSTRING", `"lkj"`, "%{QUOTEDSTRING}", `qsdklfjqsd fk"lkj"mkj`)
	go check("QUOTEDSTRING", `'lkj'`, "%{QUOTEDSTRING}", `qsdklfjqsd fk'lkj'mkj`)
	go check("QUOTEDSTRING", `'lkj'`, "%{QUOTEDSTRING}", `qsdklfjqsd fk'lkj'mkj`)
	go check("QUOTEDSTRING", `"fk'lkj'm"`, "%{QUOTEDSTRING}", `qsdklfjqsd "fk'lkj'm"kj`)
	go check("QUOTEDSTRING", `'fk"lkj"m'`, "%{QUOTEDSTRING}", `qsdklfjqsd 'fk"lkj"m'kj`)
}

func TestPatterns(t *testing.T) {
	g := NewWithConfig(&Config{SkipDefaultPatterns: true})
	if len(g.Patterns()) != 0 {
		t.Fatalf("Patterns should return 0, have '%d'", len(g.Patterns()))
	}
	name := "DAY0"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"

	g.AddPattern(name, pattern)
	g.AddPattern(name+"1", pattern)

	if len(g.Patterns()) != 2 {
		t.Fatalf("Patterns should return 2, have '%d'", len(g.Patterns()))
	}
}

func TestParseTypedWithDefaultCaptureMode(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{IPV4:ip:string} %{NUMBER:status:int} %{NUMBER:duration:float}", `127.0.0.1 200 0.8`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["ip"] != "127.0.0.1" {
			t.Fatalf("%s should be '%s' have '%s'", "ip", "127.0.0.1", captures["ip"])
		} else {
			if captures["status"] != 200 {
				t.Fatalf("%s should be '%d' have '%d'", "status", 200, captures["status"])
			} else {
				if captures["duration"] != 0.8 {
					t.Fatalf("%s should be '%d' have '%d'", "duration", 0.8, captures["duration"])
				}
			}
		}
	}
}

func TestParseTypedWithNoTypeInfo(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["timestamp"] != "23/Apr/2014:22:58:32 +0200" {
			t.Fatalf("%s should be '%s' have '%s'", "timestamp", "23/Apr/2014:22:58:32 +0200", captures["timestamp"])
		}
		if captures["TIME"] != "" {
			t.Fatalf("%s should be '%s' have '%s'", "TIME", "", captures["TIME"])
		}
	}

	g = New()
	if captures, err := g.ParseTyped("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["timestamp"] != "23/Apr/2014:22:58:32 +0200" {
			t.Fatalf("%s should be '%s' have '%s'", "timestamp", "23/Apr/2014:22:58:32 +0200", captures["timestamp"])
		}
		if captures["TIME"] != "22:58:32" {
			t.Fatalf("%s should be '%s' have '%s'", "TIME", "22:58:32", captures["TIME"])
		}
	}
}

func TestParseTypedWithIntegerTypeCoercion(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{WORD:coerced:int}", `5.75`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["coerced"] != 5 {
			t.Fatalf("%s should be '%s' have '%s'", "coerced", "5", captures["coerced"])
		}
	}
}

func TestParseTypedWithUnknownType(t *testing.T) {
	g := NewWithConfig(&Config{NamedCapturesOnly: true})
	if _, err := g.ParseTyped("%{WORD:word:unknown}", `hello`); err == nil {
		t.Fatalf("parsing an unknown type must result in a conversion error")
	}
}

func TestParseTypedErrorCaptureUnknowPattern(t *testing.T) {
	g := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.ParseTyped(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}
