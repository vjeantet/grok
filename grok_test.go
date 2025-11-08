package grok

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	g, _ := New()
	if len(g.patterns) == 0 {
		t.Fatal("the Grok object should have some patterns pre loaded")
	}

	g, _ = NewWithConfig(&Config{NamedCapturesOnly: true})
	if len(g.patterns) == 0 {
		t.Fatal("the Grok object should have some patterns pre loaded")
	}
}

func TestParseWithDefaultCaptureMode(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
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

	g, _ = New()
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
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	res, _ := g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:23:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["TIME"]) != 0 {
		t.Fatalf("DAY should be an array of 0 elements, but is '%s'", res["TIME"])
	}

	g, _ = New()
	res, _ = g.ParseToMultiMap("%{COMMONAPACHELOG} %{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:23:58:32 +0200] "GET /index.php HTTP/1.1" 404 207 127.0.0.1 - - [24/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if len(res["TIME"]) != 2 {
		t.Fatalf("TIME should be an array of 2 elements, but is '%s'", res["TIME"])
	}
	if len(res["timestamp"]) != 2 {
		t.Fatalf("timestamp should be an array of 2 elements, but is '%s'", res["timestamp"])
	}
}

func TestNewWithNoDefaultPatterns(t *testing.T) {
	g, _ := NewWithConfig(&Config{SkipDefaultPatterns: true})
	if len(g.patterns) != 0 {
		t.Fatal("Using SkipDefaultPatterns the Grok object should not have any patterns pre loaded")
	}
}

func TestAddPatternErr(t *testing.T) {
	name := "Error"
	pattern := "%{ERR}"

	g, _ := New()
	err := g.addPattern(name, pattern)
	if err == nil {
		t.Fatalf("AddPattern should returns an error when path is invalid")
	}
}

func TestAddPatternsFromPathErr(t *testing.T) {
	g, _ := New()
	err := g.AddPatternsFromPath("./Lorem ipsum Minim qui in.")
	if err == nil {
		t.Fatalf("AddPatternsFromPath should returns an error when path is invalid")
	}
}

func TestConfigPatternsDir(t *testing.T) {
	g, err := NewWithConfig(&Config{PatternsDir: []string{"./patterns"}})
	if err != nil {
		t.Error(err)
	}

	if captures, err := g.Parse("%{SYSLOGLINE}", `Sep 12 23:19:02 docker syslog-ng[25389]: syslog-ng starting up; version='3.5.3'`); err != nil {
		t.Fatalf("error : %s", err.Error())
	} else {
		// pp.Print(captures)
		if captures["program"] != "syslog-ng" {
			t.Fatalf("%s should be '%s' have '%s'", "program", "syslog-ng", captures["program"])
		}
	}

}

func TestAddPatternsFromPathFileOpenErr(t *testing.T) {
	t.Skipped()
}

func TestAddPatternsFromPathFile(t *testing.T) {
	g, _ := New()
	err := g.AddPatternsFromPath("./patterns/grok-patterns")
	if err != nil {
		t.Fatalf("err %#v", err)
	}
}

func TestAddPattern(t *testing.T) {
	name := "DAYO"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"

	g, _ := New()
	cPatterns := len(g.patterns)
	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)
	if len(g.patterns) != cPatterns+2 {
		t.Fatalf("%d Default patterns should be available, have %d", cPatterns+2, len(g.patterns))
	}

	g, _ = NewWithConfig(&Config{NamedCapturesOnly: true})
	cPatterns = len(g.patterns)
	g.AddPattern(name, pattern)
	g.AddPattern(name+"2", pattern)
	if len(g.patterns) != cPatterns+2 {
		t.Fatalf("%d NamedCapture patterns should be available, have %d", cPatterns+2, len(g.patterns))
	}
}

func TestMatch(t *testing.T) {
	g, _ := New()
	g.AddPatternsFromPath("./patterns")

	if r, err := g.Match("%{MONTH}", "June"); !r {
		t.Fatalf("June should match %s: err=%s", "%{MONTH}", err.Error())
	}

}
func TestDoesNotMatch(t *testing.T) {
	g, _ := New()
	g.AddPatternsFromPath("./patterns")
	if r, _ := g.Match("%{MONTH}", "13"); r {
		t.Fatalf("13 should not match %s", "%{MONTH}")
	}
}

func TestErrorMatch(t *testing.T) {
	g, _ := New()
	if _, err := g.Match("(", "13"); err == nil {
		t.Fatal("Error expected")
	}

}

func TestShortName(t *testing.T) {
	g, _ := New()
	g.AddPattern("A", "a")

	m, err := g.Match("%{A}", "a")
	if err != nil {
		t.Fatalf("a should match %%{A}: err=%s", err.Error())
	}
	if !m {
		t.Fatal("%%{A} didn't match 'a'")
	}
}

func TestDayCompile(t *testing.T) {
	g, _ := New()
	g.AddPattern("DAY", "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)")
	pattern := "%{DAY}"
	_, err := g.compile(pattern)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestErrorCompile(t *testing.T) {
	g, _ := New()
	_, err := g.compile("(")
	if err == nil {
		t.Fatal("Error:", err)
	}
}

func TestNamedCaptures(t *testing.T) {
	g, _ := New()
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

	check("day-of.week", "Tue",
		"%{DAY:day-of.week}",
		"Tue May 15 11:21:42 [conn1047685] moveChunk deleted: 7157",
	)
}

func TestErrorCaptureUnknowPattern(t *testing.T) {
	g, _ := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.Parse(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestErrorCaptureInvalidPattern(t *testing.T) {
	g, _ := New()
	pattern := "%{-InvalidPattern-}"
	expectederr := "invalid pattern %{-InvalidPattern-}"
	_, err := g.Parse(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
	if err.Error() != expectederr {
		t.Fatalf("Expected error %q but got %q", expectederr, err.Error())
	}
}

func TestParse(t *testing.T) {
	g, _ := New()
	g.AddPatternsFromPath("./patterns")
	res, _ := g.Parse("%{DAY}", "Tue qds")
	if res["DAY"] != "Tue" {
		t.Fatalf("DAY should be 'Tue' have '%s'", res["DAY"])
	}
}

func TestErrorParseToMultiMap(t *testing.T) {
	g, _ := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.ParseToMultiMap(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParseToMultiMap(t *testing.T) {
	g, _ := New()
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
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
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
	g, _ := New()
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
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
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

func TestRemoveEmptyValues(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true, RemoveEmptyValues: true})

	capturesExists := func(key, pattern, text string) {
		if captures, err := g.Parse(pattern, text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if _, ok := captures[key]; ok {
				t.Fatalf("%s should be absent", key)
			}
		}
	}

	capturesExists("rawrequest", "%{COMMONAPACHELOG}",
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
	)

}

func TestCapturesAndNamedCapture(t *testing.T) {

	check := func(key, value, pattern, text string) {
		g, _ := New()
		if captures, err := g.Parse(pattern, text); err != nil {
			t.Fatalf("error can not capture : %s", err.Error())
		} else {
			if captures[key] != value {
				t.Fatalf("%s should be '%s' have '%s'", key, value, captures[key])
			}
		}
	}

	checkNamed := func(key, value, pattern, text string) {
		g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
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
	g, _ := New()
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
	g, _ := NewWithConfig(&Config{SkipDefaultPatterns: true})
	if len(g.patterns) != 0 {
		t.Fatalf("Patterns should return 0, have '%d'", len(g.patterns))
	}
	name := "DAY0"
	pattern := "(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)"

	g.AddPattern(name, pattern)
	g.AddPattern(name+"1", pattern)
	if len(g.patterns) != 2 {
		t.Fatalf("Patterns should return 2, have '%d'", len(g.patterns))
	}
}

func TestParseTypedWithDefaultCaptureMode(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
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
					t.Fatalf("%s should be '%f' have '%f'", "duration", 0.8, captures["duration"])
				}
			}
		}
	}
}

func TestParseTypedWithNoTypeInfo(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["timestamp"] != "23/Apr/2014:22:58:32 +0200" {
			t.Fatalf("%s should be '%s' have '%s'", "timestamp", "23/Apr/2014:22:58:32 +0200", captures["timestamp"])
		}
		if captures["TIME"] != nil {
			t.Fatalf("%s should be nil have '%s'", "TIME", captures["TIME"])
		}
	}

	g, _ = New()
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
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{WORD:coerced:int}", `5.75`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["coerced"] != 5 {
			t.Fatalf("%s should be '%s' have '%s'", "coerced", "5", captures["coerced"])
		}
	}
}

func TestParseTypedWithUnknownType(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	if _, err := g.ParseTyped("%{WORD:word:unknown}", `hello`); err == nil {
		t.Fatalf("parsing an unknown type must result in a conversion error")
	}
}

func TestParseTypedErrorCaptureUnknowPattern(t *testing.T) {
	g, _ := New()
	pattern := "%{UNKNOWPATTERN}"
	_, err := g.ParseTyped(pattern, "")
	if err == nil {
		t.Fatal("Expected error not set")
	}
}

func TestParseTypedWithTypedParents(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	g.AddPattern("TESTCOMMON", `%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes:int}|-)`)
	if captures, err := g.ParseTyped("%{TESTCOMMON}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["bytes"] != 207 {
			t.Fatalf("%s should be '%s' have '%s'", "bytes", "207", captures["bytes"])
		}
	}
}

func TestParseTypedWithSemanticHomonyms(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true, SkipDefaultPatterns: true})

	g.AddPattern("BASE10NUM", `([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)`)
	g.AddPattern("NUMBER", `(?:%{BASE10NUM})`)
	g.AddPattern("MYNUM", `%{NUMBER:bytes:int}`)
	g.AddPattern("MYSTR", `%{NUMBER:bytes:string}`)

	if captures, err := g.ParseTyped("%{MYNUM}", `207`); err != nil {
		t.Fatalf("error can not scapture : %s", err.Error())
	} else {
		if captures["bytes"] != 207 {
			t.Fatalf("%s should be %#v have %#v", "bytes", 207, captures["bytes"])
		}
	}
	if captures, err := g.ParseTyped("%{MYSTR}", `207`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["bytes"] != "207" {
			t.Fatalf("%s should be %#v have %#v", "bytes", "207", captures["bytes"])
		}
	}
}

func TestParseTypedWithAlias(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{NUMBER:access.response_code:int}", `404`); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		if captures["access.response_code"] != 404 {
			t.Fatalf("%s should be %#v have %#v", "access.response_code", 404, captures["access.response_code"])
		}
	}
}

var resultNew *Grok

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var g *Grok
	// run the check function b.N times
	for n := 0; n < b.N; n++ {
		g, _ = NewWithConfig(&Config{NamedCapturesOnly: true})
	}
	resultNew = g
}

func BenchmarkCaptures(b *testing.B) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	b.ReportAllocs()
	b.ResetTimer()
	// run the check function b.N times
	for n := 0; n < b.N; n++ {
		g.Parse(`%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)`, `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	}
}

func BenchmarkCapturesTypedFake(b *testing.B) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	b.ReportAllocs()
	b.ResetTimer()
	// run the check function b.N times
	for n := 0; n < b.N; n++ {
		g.Parse(`%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)`, `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	}
}

func BenchmarkCapturesTypedReal(b *testing.B) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	b.ReportAllocs()
	b.ResetTimer()
	// run the check function b.N times
	for n := 0; n < b.N; n++ {
		g.ParseTyped(`%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion:int})?|%{DATA:rawrequest})" %{NUMBER:response:int} (?:%{NUMBER:bytes:int}|-)`, `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	}
}

func BenchmarkParallelCaptures(b *testing.B) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			g.Parse(`%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)`, `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
		}
	})
}

func TestGrok_AddPatternsFromMap_not_exist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddPatternsFromMap panics: %v", r)
		}
	}()
	g, _ := NewWithConfig(&Config{SkipDefaultPatterns: true})
	err := g.AddPatternsFromMap(map[string]string{
		"SOME": "%{NOT_EXIST}",
	})
	if err == nil {
		t.Errorf("AddPatternsFromMap should returns an error")
	}
}

func TestGrok_AddPatternsFromMap_simple(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddPatternsFromMap panics: %v", r)
		}
	}()
	g, _ := NewWithConfig(&Config{SkipDefaultPatterns: true})
	err := g.AddPatternsFromMap(map[string]string{
		"NO3": `\d{3}`,
	})
	if err != nil {
		t.Errorf("AddPatternsFromMap returns an error: %v", err)
	}
	mss, err := g.Parse("%{NO3:match}", "333")
	if err != nil {
		t.Error("parsing error:", err)
		t.FailNow()
	}
	if mss["match"] != "333" {
		t.Errorf("bad match: expected 333, got %s", mss["match"])
	}
}

func TestGrok_AddPatternsFromMap_complex(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddPatternsFromMap panics: %v", r)
		}
	}()
	g, _ := NewWithConfig(&Config{
		SkipDefaultPatterns: true,
		NamedCapturesOnly:   true,
	})
	err := g.AddPatternsFromMap(map[string]string{
		"NO3": `\d{3}`,
		"NO6": "%{NO3}%{NO3}",
	})
	if err != nil {
		t.Errorf("AddPatternsFromMap returns an error: %v", err)
	}
	mss, err := g.Parse("%{NO6:number}", "333666")
	if err != nil {
		t.Error("parsing error:", err)
		t.FailNow()
	}
	if mss["number"] != "333666" {
		t.Errorf("bad match: expected 333666, got %s", mss["match"])
	}
}

func TestParseStream(t *testing.T) {
	g, _ := New()
	pTest := func(m map[string]string) error {
		ts, ok := m["timestamp"]
		if !ok {
			t.Error("timestamp not found")
		}
		if len(ts) == 0 {
			t.Error("empty timestamp")
		}
		return nil
	}
	const testLog = `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207
127.0.0.1 - - [23/Apr/2014:22:59:32 +0200] "GET /index.php HTTP/1.1" 404 207
127.0.0.1 - - [23/Apr/2014:23:00:32 +0200] "GET /index.php HTTP/1.1" 404 207
`

	r := bufio.NewReader(strings.NewReader(testLog))
	if err := g.ParseStream(r, "%{COMMONAPACHELOG}", pTest); err != nil {
		t.Fatal(err)
	}
}

func TestParseStreamError(t *testing.T) {
	g, _ := New()
	pTest := func(m map[string]string) error {
		if _, ok := m["timestamp"]; !ok {
			return fmt.Errorf("timestamp not found")
		}
		return nil
	}
	const testLog = `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207
127.0.0.1 - - [xxxxxxxxxxxxxxxxxxxx +0200] "GET /index.php HTTP/1.1" 404 207
127.0.0.1 - - [23/Apr/2014:23:00:32 +0200] "GET /index.php HTTP/1.1" 404 207
`

	r := bufio.NewReader(strings.NewReader(testLog))
	if err := g.ParseStream(r, "%{COMMONAPACHELOG}", pTest); err == nil {
		t.Fatal("Error expected")
	}
}

func TestParseStreamCompileError(t *testing.T) {
	g, _ := New()
	pTest := func(m map[string]string) error {
		return nil
	}
	r := bufio.NewReader(strings.NewReader("test"))
	if err := g.ParseStream(r, "%{UNKNOWNPATTERN}", pTest); err == nil {
		t.Fatal("Error expected when pattern cannot be compiled")
	}
}

func TestNewWithConfigWithInvalidPatterns(t *testing.T) {
	_, err := NewWithConfig(&Config{
		Patterns: map[string]string{
			"INVALID": "%{NONEXISTENT}",
		},
	})
	if err == nil {
		t.Fatal("Error expected when config contains invalid patterns")
	}
}

func TestNewWithConfigWithInvalidPatternsDir(t *testing.T) {
	_, err := NewWithConfig(&Config{
		PatternsDir: []string{"./nonexistent_directory"},
	})
	if err == nil {
		t.Fatal("Error expected when PatternsDir contains invalid path")
	}
}

func TestParseTypedWithRemoveEmptyValues(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true, RemoveEmptyValues: true})

	captures, err := g.ParseTyped("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	}

	if _, exists := captures["rawrequest"]; exists {
		t.Fatal("rawrequest should not exist when RemoveEmptyValues is true")
	}
}

func TestParseToMultiMapWithRemoveEmptyValues(t *testing.T) {
	g, _ := NewWithConfig(&Config{RemoveEmptyValues: true})

	res, err := g.ParseToMultiMap("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if err != nil {
		t.Fatalf("error can not parse : %s", err.Error())
	}

	if _, exists := res["rawrequest"]; exists {
		t.Fatal("rawrequest should not exist when RemoveEmptyValues is true")
	}
}

func TestAddPatternsFromMapWithInvalidPatternSyntax(t *testing.T) {
	g, _ := NewWithConfig(&Config{SkipDefaultPatterns: true})

	err := g.AddPatternsFromMap(map[string]string{
		"VALID":   `\d+`,
		"INVALID": "%{-INVALID}",
	})

	if err == nil {
		t.Fatal("Error expected when pattern contains invalid syntax")
	}
}

func TestParseWithRemoveEmptyValues(t *testing.T) {
	g, _ := NewWithConfig(&Config{RemoveEmptyValues: true})

	captures, err := g.Parse("%{COMMONAPACHELOG}", `127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
	if err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	}

	// rawrequest should be empty in this log line and should be removed
	if _, exists := captures["rawrequest"]; exists {
		t.Fatal("rawrequest should not exist when RemoveEmptyValues is true and value is empty")
	}

	// Verify non-empty values are still present
	if captures["clientip"] != "127.0.0.1" {
		t.Fatal("clientip should still be present")
	}
}

func TestGreedyDataWithMixedRegex(t *testing.T) {
	g, _ := New()

	pattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND`

	// Debug: get the compiled regex to see what it looks like
	gr, err := g.compile(pattern)
	if err != nil {
		t.Fatalf("error compiling: %s", err.Error())
	}
	t.Logf("Compiled regex: %s", gr.regexp.String())
	t.Logf("SubexpNames: %v", gr.regexp.SubexpNames())

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("error parsing: %s", err.Error())
	}

	t.Logf("Captures: %+v", captures)
	t.Logf("Number of captures: %d", len(captures))

	if len(captures) == 0 {
		t.Fatal("Expected non-empty captures but got empty map")
	}

	if captures["var1"] != "/opt/facs/casrepos/fa/common.jar" {
		t.Fatalf("var1 should be '/opt/facs/casrepos/fa/common.jar' but got '%s'", captures["var1"])
	}

	if captures["var2"] != "44228-9915812-0" {
		t.Fatalf("var2 should be '44228-9915812-0' but got '%s'", captures["var2"])
	}

	if captures["var3"] != "FOUND" {
		t.Fatalf("var3 should be 'FOUND' but got '%s'", captures["var3"])
	}
}

func TestGreedyDataWithEmptyValues(t *testing.T) {
	g, _ := New()

	pattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`

	// Test case that might produce empty values
	text := ": FOUND"

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("error parsing: %s", err.Error())
	}

	t.Logf("Captures for '%s': %+v", text, captures)
	t.Logf("Number of captures: %d", len(captures))

	// This should return empty map since the pattern doesn't match
	if len(captures) != 0 {
		t.Logf("Warning: Expected empty map for input '%s' but got %+v", text, captures)
	}
}

func TestGreedyDataCanMatchEmpty(t *testing.T) {
	g, _ := New()

	//  Test if GREEDYDATA (.*) can match empty strings
	pattern := `%{GREEDYDATA:before}FOUND%{GREEDYDATA:after}`
	text := "FOUND"

	gr, _ := g.compile(pattern)
	t.Logf("Pattern: %s", gr.regexp.String())

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("error parsing: %s", err.Error())
	}

	t.Logf("Captures: %+v", captures)

	// .* can match empty string, so this should work
	if captures["before"] != "" {
		t.Fatalf("Expected 'before' to be empty but got '%s'", captures["before"])
	}
	if captures["after"] != "" {
		t.Fatalf("Expected 'after' to be empty but got '%s'", captures["after"])
	}
}

func TestGreedyDataWithTrailingNewline(t *testing.T) {
	g, _ := New()

	pattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`

	// Test with trailing newline - .* doesn't match newlines by default
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND
`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("error parsing: %s", err.Error())
	}

	t.Logf("Text (with newline): %q", text)
	t.Logf("Captures: %+v", captures)
	t.Logf("Number of captures: %d", len(captures))

	// THIS MIGHT BE THE ISSUE! .* doesn't match newlines
	if len(captures) == 0 {
		t.Log("FOUND IT! Pattern returns empty map when input has trailing newline")
		t.Log("This is because .* doesn't match newline characters")
	}
}

func TestGreedyDataWithRemoveEmptyValues(t *testing.T) {
	// Test if RemoveEmptyValues causes issues
	g, _ := NewWithConfig(&Config{RemoveEmptyValues: true})

	pattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("error parsing: %s", err.Error())
	}

	t.Logf("Captures (RemoveEmptyValues=true): %+v", captures)
	t.Logf("Number of captures: %d", len(captures))

	// Should still have 3 captures since none are empty
	if len(captures) != 3 {
		t.Fatalf("Expected 3 captures but got %d: %+v", len(captures), captures)
	}

	// Now test with input that might produce empty captures
	text2 := ": FOUND"
	captures2, _ := g.Parse(pattern, text2)
	t.Logf("Captures for '%s' (RemoveEmptyValues=true): %+v", text2, captures2)
}

func TestDATAvsGREEDYDATA(t *testing.T) {
	g, _ := New()

	// Compare DATA (.*?) non-greedy vs GREEDYDATA (.*) greedy
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND`

	// Test with GREEDYDATA (greedy)
	greedyPattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`
	greedyCaptures, _ := g.Parse(greedyPattern, text)
	t.Logf("GREEDYDATA pattern: %s", greedyPattern)
	t.Logf("GREEDYDATA captures: %+v", greedyCaptures)

	// Test with DATA (non-greedy)
	dataPattern := `%{DATA:var1}: %{DATA:var2} (?<var3>FOUND)`
	dataCaptures, _ := g.Parse(dataPattern, text)
	t.Logf("DATA pattern: %s", dataPattern)
	t.Logf("DATA captures: %+v", dataCaptures)

	// Both should work but may capture differently
	if len(greedyCaptures) == 0 {
		t.Error("GREEDYDATA returned empty map!")
	}
	if len(dataCaptures) == 0 {
		t.Error("DATA returned empty map!")
	}

	// Check the compiled patterns to see the difference
	grGREEDY, _ := g.compile(greedyPattern)
	grDATA, _ := g.compile(dataPattern)
	t.Logf("GREEDYDATA compiled: %s", grGREEDY.regexp.String())
	t.Logf("DATA compiled: %s", grDATA.regexp.String())
}

// TestIssue_GrokParseEmptyMap is a regression test for the reported issue where
// Grok.Parse with GREEDYDATA pattern and mixed regex syntax returns empty map.
// This test documents the expected behavior.
func TestIssue_GrokParseEmptyMap(t *testing.T) {
	g, err := New()
	if err != nil {
		t.Fatalf("Failed to create Grok instance: %v", err)
	}

	// The exact pattern and text from the issue report
	pattern := `%{GREEDYDATA:var1}: %{GREEDYDATA:var2} (?<var3>FOUND)`
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND`

	// This should work and return 3 captures
	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Verify we got results
	if len(captures) == 0 {
		t.Fatal("BUG REPRODUCED: Got empty map!")
	}

	// Verify expected values
	expectedCaptures := map[string]string{
		"var1": "/opt/facs/casrepos/fa/common.jar",
		"var2": "44228-9915812-0",
		"var3": "FOUND",
	}

	for key, expected := range expectedCaptures {
		if actual, ok := captures[key]; !ok {
			t.Errorf("Missing capture %q", key)
		} else if actual != expected {
			t.Errorf("Capture %q: expected %q, got %q", key, expected, actual)
		}
	}

	t.Logf("✓ Pattern works correctly")
	t.Logf("  Pattern: %s", pattern)
	t.Logf("  Text: %s", text)
	t.Logf("  Captures: %+v", captures)
}

func TestFieldNamesWithParentheses(t *testing.T) {
	g, _ := New()

	// Test IIS log pattern with parentheses in field names like cs(User-Agent) and cs(Referer)
	pattern := "%{TIMESTAMP_ISO8601:logtime} %{WORD:s-sitename} %{WORD:s-computername} %{IPORHOST:s-ip} %{WORD:cs-method} %{NOTSPACE:cs-uri-stem} %{NOTSPACE:cs-uri-query} %{NUMBER:s-port} %{NOTSPACE:cs-username} %{IPORHOST:c-ip} %{NOTSPACE:cs-version} %{NOTSPACE:cs(User-Agent)} %{NOTSPACE:cs(Referer)} %{IPORHOST:cs-host} %{NUMBER:sc-status} %{NUMBER:sc-substatus} %{NUMBER:sc-win32-status} %{NUMBER:sc-bytes} %{NUMBER:cs-bytes} %{NUMBER:time-taken}"
	logLine := "2018-02-02 00:01:32 W3SVC1 UKAPPSVR 172.18.131.173 GET /123/I/Home/PLMonstants - 80 Joe+Bloggs 172.18.17.185 HTTP/1.1 Mozilla/5.0+(Windows+NT+6.1;+Trident/7.0;+rv:11.0)+like+Gecko https://blahblah.co.uk/theappname/live/app/thingy localhost 200 0 0 3393 2644 90"

	captures, err := g.Parse(pattern, logLine)
	if err != nil {
		t.Fatalf("Failed to parse IIS log with parentheses in field names: %s", err.Error())
	}

	// Verify field names with parentheses are captured correctly
	if captures["cs(User-Agent)"] != "Mozilla/5.0+(Windows+NT+6.1;+Trident/7.0;+rv:11.0)+like+Gecko" {
		t.Fatalf("cs(User-Agent) should be 'Mozilla/5.0+(Windows+NT+6.1;+Trident/7.0;+rv:11.0)+like+Gecko' but got '%s'", captures["cs(User-Agent)"])
	}

	if captures["cs(Referer)"] != "https://blahblah.co.uk/theappname/live/app/thingy" {
		t.Fatalf("cs(Referer) should be 'https://blahblah.co.uk/theappname/live/app/thingy' but got '%s'", captures["cs(Referer)"])
	}

	// Verify other fields still work correctly
	if captures["logtime"] != "2018-02-02 00:01:32" {
		t.Fatalf("logtime should be '2018-02-02 00:01:32' but got '%s'", captures["logtime"])
	}

	if captures["s-sitename"] != "W3SVC1" {
		t.Fatalf("s-sitename should be 'W3SVC1' but got '%s'", captures["s-sitename"])
	}

	if captures["cs-method"] != "GET" {
		t.Fatalf("cs-method should be 'GET' but got '%s'", captures["cs-method"])
	}

	if captures["c-ip"] != "172.18.17.185" {
		t.Fatalf("c-ip should be '172.18.17.185' but got '%s'", captures["c-ip"])
	}

	if captures["sc-status"] != "200" {
		t.Fatalf("sc-status should be '200' but got '%s'", captures["sc-status"])
	}
}

func TestNestedFieldsWithBrackets(t *testing.T) {
	g, _ := New()

	// Test nested field notation using brackets
	pattern := `%{GREEDYDATA:[field1][nestedField1]}: %{GREEDYDATA:[field2][nestedField2]}`
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse with nested field brackets: %s", err.Error())
	}

	if captures["[field1][nestedField1]"] != "/opt/facs/casrepos/fa/common.jar" {
		t.Fatalf("[field1][nestedField1] should be '/opt/facs/casrepos/fa/common.jar' but got '%s'", captures["[field1][nestedField1]"])
	}

	if captures["[field2][nestedField2]"] != "44228-9915812-0" {
		t.Fatalf("[field2][nestedField2] should be '44228-9915812-0' but got '%s'", captures["[field2][nestedField2]"])
	}
}

func TestNestedFieldsWithNativeRegex(t *testing.T) {
	g, _ := New()

	// Test native regex syntax with bracket notation
	pattern := `(?<[field1][nested1]>\S+): (?<[field2][nested2]>\S+)`
	text := `/opt/facs/common.jar: 12345`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse with native regex bracket notation: %s", err.Error())
	}

	if captures["[field1][nested1]"] != "/opt/facs/common.jar" {
		t.Fatalf("[field1][nested1] should be '/opt/facs/common.jar' but got '%s'", captures["[field1][nested1]"])
	}

	if captures["[field2][nested2]"] != "12345" {
		t.Fatalf("[field2][nested2] should be '12345' but got '%s'", captures["[field2][nested2]"])
	}
}

func TestNestedFieldsMixed(t *testing.T) {
	g, _ := New()

	// Test mixed syntax - Grok patterns with brackets and native regex with brackets
	pattern := `%{GREEDYDATA:[field1][nestedField1]}: %{GREEDYDATA:[field2][nestedField2]} (?<[field3][nestedField3]>FOUND)`
	text := `/opt/facs/casrepos/fa/common.jar: 44228-9915812-0 FOUND`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse with mixed bracket notation: %s", err.Error())
	}

	if captures["[field1][nestedField1]"] != "/opt/facs/casrepos/fa/common.jar" {
		t.Fatalf("[field1][nestedField1] should be '/opt/facs/casrepos/fa/common.jar' but got '%s'", captures["[field1][nestedField1]"])
	}

	if captures["[field2][nestedField2]"] != "44228-9915812-0" {
		t.Fatalf("[field2][nestedField2] should be '44228-9915812-0' but got '%s'", captures["[field2][nestedField2]"])
	}

	if captures["[field3][nestedField3]"] != "FOUND" {
		t.Fatalf("[field3][nestedField3] should be 'FOUND' but got '%s'", captures["[field3][nestedField3]"])
	}
}

func TestNestedFieldsMultipleLevels(t *testing.T) {
	g, _ := New()

	// Test multiple levels of nesting
	pattern := `%{WORD:[level1][level2][level3]}: %{NUMBER:[a][b][c][d]}`
	text := `test: 12345`

	captures, err := g.Parse(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse with multiple nesting levels: %s", err.Error())
	}

	if captures["[level1][level2][level3]"] != "test" {
		t.Fatalf("[level1][level2][level3] should be 'test' but got '%s'", captures["[level1][level2][level3]"])
	}

	if captures["[a][b][c][d]"] != "12345" {
		t.Fatalf("[a][b][c][d] should be '12345' but got '%s'", captures["[a][b][c][d]"])
	}
}

func TestNestedFieldsTyped(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})

	// Test nested fields with type coercion
	pattern := `%{NUMBER:[server][port]:int} %{NUMBER:[stats][count]:float}`
	text := `8080 123.45`

	captures, err := g.ParseTyped(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse nested fields with types: %s", err.Error())
	}

	// ParseTyped returns nested maps
	server, ok := captures["server"].(map[string]interface{})
	if !ok {
		t.Fatalf("server should be a nested map but got %T", captures["server"])
	}

	if server["port"] != 8080 {
		t.Fatalf("server.port should be 8080 but got %v", server["port"])
	}

	stats, ok := captures["stats"].(map[string]interface{})
	if !ok {
		t.Fatalf("stats should be a nested map but got %T", captures["stats"])
	}

	if stats["count"] != 123.45 {
		t.Fatalf("stats.count should be 123.45 but got %v", stats["count"])
	}
}

func TestNestedFieldsToMultiMap(t *testing.T) {
	g, _ := New()

	// Test nested fields with ParseToMultiMap
	pattern := `%{WORD:[field][nested]} %{WORD:[field][nested]}`
	text := `first second`

	captures, err := g.ParseToMultiMap(pattern, text)
	if err != nil {
		t.Fatalf("Failed to parse nested fields to multimap: %s", err.Error())
	}

	if len(captures["[field][nested]"]) != 2 {
		t.Fatalf("[field][nested] should have 2 values but got %d", len(captures["[field][nested]"]))
	}

	if captures["[field][nested]"][0] != "first" {
		t.Fatalf("[field][nested][0] should be 'first' but got '%s'", captures["[field][nested]"][0])
	}

	if captures["[field][nested]"][1] != "second" {
		t.Fatalf("[field][nested][1] should be 'second' but got '%s'", captures["[field][nested]"][1])
	}
}

func TestParseTypedWithNested(t *testing.T) {
	g, _ := NewWithConfig(&Config{NamedCapturesOnly: true})
	if captures, err := g.ParseTyped("%{TIMESTAMP_ISO8601:time} %{USER:[user][name]}@%{HOSTNAME:[user][host]} %{WORD:action} %{POSINT:[net][bytes]:int} bytes from %{IP:[net][source][ip]}:%{POSINT:[net][source][port]:int}", "2023-04-08T11:55:00+0200 john.doe@example.com send 230 bytes from 198.51.100.65:2342"); err != nil {
		t.Fatalf("error can not capture : %s", err.Error())
	} else {
		expected := map[string]interface{}{
			"time":   "2023-04-08T11:55:00+0200",
			"action": "send",
			"user": map[string]interface{}{
				"name": "john.doe",
				"host": "example.com",
			},
			"net": map[string]interface{}{
				"bytes": 230,
				"source": map[string]interface{}{
					"ip":   "198.51.100.65",
					"port": 2342,
				},
			},
		}
		if fmt.Sprint(expected) != fmt.Sprint(captures) {
			t.Fatalf("Expected nested map: %s got %s", expected, captures)
		}
	}
}
