//go:generate stringer -type=Option
//go:generate patternstoregex

package grok

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

//Option are used as params with Parse() and ParseToMultiMap()
type Option uint

const (
	DEFAULTCAPTURE    Option = iota // Parse() And ParseToMulti() will return every group
	NAMEDCAPTURE                    // Parse() And ParseToMulti() will return only named captures
	NODEFAULTPATTERNS               // Do not use default patterns compiled in the lib
)

// Grok Type
type Grok struct {
	compiledPattern    map[Option]map[string]*regexp.Regexp
	patterns           map[Option]map[string]string
	serviceMu          sync.Mutex
	defaultCaptureMode Option
}

// New returns a Grok struct
// Options availables :
// grok.NODEFAULTPATTERNS => Do not use compiled-in patterns
// grok.DEFAULTCAPTURE    => Parse and ParseToMulti will return all groups
// grok.NAMEDCAPTURE 	  => Parse and ParseToMulti will return only named captures
// Exemple
// g := grok.New()
// g := grok.New(grok.NAMEDCAPTURE)
// g := grok.New(grok.NAMEDCAPTURE, grok.NODEFAULTPATTERNS)
// g := grok.New(grok.NODEFAULTPATTERNS, grok.DEFAULTCAPTURE)
func New(opt ...Option) *Grok {
	o := new(Grok)
	o.patterns = patterns
	o.defaultCaptureMode = DEFAULTCAPTURE
	o.compiledPattern = map[Option]map[string]*regexp.Regexp{
		DEFAULTCAPTURE: map[string]*regexp.Regexp{},
		NAMEDCAPTURE:   map[string]*regexp.Regexp{},
	}

	for _, v := range opt {
		if v == NODEFAULTPATTERNS {
			o.patterns = map[Option]map[string]string{
				DEFAULTCAPTURE: map[string]string{},
				NAMEDCAPTURE:   map[string]string{},
			}
		}
		if v == NAMEDCAPTURE {
			o.defaultCaptureMode = NAMEDCAPTURE
		}
	}

	return o
}

// AddPattern add a pattern to grok
func (g *Grok) AddPattern(name string, pattern string) {
	p1, p2 := denormalizePattern(pattern, g.patterns)
	g.patterns[DEFAULTCAPTURE][name] = p1
	g.patterns[NAMEDCAPTURE][name] = p2
}

func (g *Grok) cache(pattern string, cr *regexp.Regexp, kindOfCapture Option) {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()
	g.compiledPattern[kindOfCapture][pattern] = cr
}

func (g *Grok) cacheExists(pattern string, kindOfCapture Option) bool {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()

	if _, ok := g.compiledPattern[kindOfCapture][pattern]; ok {
		return true
	}

	return false
}

func (g *Grok) compile(pattern string, kindOfCapture Option) (*regexp.Regexp, error) {
	if g.cacheExists(pattern, kindOfCapture) {
		return g.compiledPattern[kindOfCapture][pattern], nil
	}

	//search for %{...:...}
	r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
	newPattern := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		customname := names[0]
		if len(names) != 1 {
			customname = names[1]
		}
		//search for replacements
		if ok := g.patterns[kindOfCapture][names[0]]; ok == "" {
			return nil, fmt.Errorf("ERROR no pattern found for %%{%s}", names[0])
		}
		replace := fmt.Sprintf("(?P<%s>%s)", customname, g.patterns[kindOfCapture][names[0]])

		//build the new regexp
		newPattern = strings.Replace(newPattern, values[0], replace, -1)
	}
	patternCompiled, err := regexp.Compile(newPattern)

	if err != nil {
		return nil, err
	}
	g.cache(pattern, patternCompiled, kindOfCapture)
	return patternCompiled, nil

}

// Match returns true when text match the compileed pattern
func (g *Grok) Match(pattern, text string) (bool, error) {
	cr, err := g.compile(pattern, DEFAULTCAPTURE)

	if err != nil {
		return false, err
	}

	if m := cr.MatchString(text); !m {
		return false, nil
	}

	return true, nil
}

// Parse returns a string map with captured string based on provided pattern over the text
func (g *Grok) Parse(pattern string, text string, options ...Option) (map[string]string, error) {
	var kindOfCapture Option
	kindOfCapture = g.defaultCaptureMode
	for _, v := range options {
		if v == NAMEDCAPTURE {
			kindOfCapture = NAMEDCAPTURE
		}
		if v == DEFAULTCAPTURE {
			kindOfCapture = DEFAULTCAPTURE
		}
	}

	return g.parse(pattern, text, kindOfCapture)
}

// Parse returns a string map with captured string (only named or all) based on provided pattern over the text
func (g *Grok) parse(pattern string, text string, kindOfCapture Option) (map[string]string, error) {
	cr, err := g.compile(pattern, kindOfCapture)
	if err != nil {
		return nil, err
	}

	match := cr.FindStringSubmatch(text)
	captures := make(map[string]string)
	if len(match) > 0 {
		for i, name := range cr.SubexpNames() {
			captures[name] = match[i]
		}
	}

	return captures, nil
}

// ParseToMultiMap works just like Parse, except that it allows to map multiple values to the same capture name.
func (g *Grok) ParseToMultiMap(pattern string, text string, options ...Option) (map[string][]string, error) {
	var kindOfCapture Option
	kindOfCapture = g.defaultCaptureMode
	for _, v := range options {
		if v == NAMEDCAPTURE {
			kindOfCapture = NAMEDCAPTURE
		}
		if v == DEFAULTCAPTURE {
			kindOfCapture = DEFAULTCAPTURE
		}
	}

	return g.parseToMultiMap(pattern, text, kindOfCapture)
}

// parseToMultiMap works just like Parse, except that it allows to map multiple values to the same capture name.
func (g *Grok) parseToMultiMap(pattern string, text string, kindOfCapture Option) (map[string][]string, error) {
	multiCaptures := make(map[string][]string)
	cr, err := g.compile(pattern, kindOfCapture)
	if err != nil {
		return nil, err
	}

	match := cr.FindStringSubmatch(text)
	if len(match) > 0 {
		for i, name := range cr.SubexpNames() {
			multiCaptures[name] = append(multiCaptures[name], match[i])
		}
	}

	return multiCaptures, nil
}

// AddPatternsFromPath loads grok patterns from a file or files from a directory
func (g *Grok) AddPatternsFromPath(path string) error {

	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path = path + "/*"
		}
	} else {
		return fmt.Errorf("invalid path : %s", path)
	}

	var patternDependancies = graph{}
	var fileContent = map[string]string{}

	//List all files if path folder
	files, _ := filepath.Glob(path)
	for _, file := range files {
		inFile, _ := os.Open(file)

		reader := bufio.NewReader(inFile)
		scanner := bufio.NewScanner(reader)
		r, _ := regexp.Compile(`%{(\w+):?(\w+)?}`)

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 { // line has text
				if l[0] != '#' { // line does not start with #
					names := strings.SplitN(l, " ", 2)
					// names[0] = key
					// names[1] = pattern
					fileContent[names[0]] = names[1]

					keys := []string{}
					for _, key := range r.FindAllStringSubmatch(names[1], -1) {
						keys = append(keys, key[1])
					}
					patternDependancies[names[0]] = keys
				}
			}
		}
		inFile.Close()
	}

	order, _ := sortGraph(patternDependancies)
	order = reverseList(order)

	for _, key := range order {
		g.AddPattern(key, fileContent[key])
	}

	return nil
}

func denormalizePattern(pattern string, finalPatterns map[Option]map[string]string) (string, string) {
	r, _ := regexp.Compile(`%{((\w+):?(\w+)?)}`)
	newPatternNC := pattern
	newPatternDF := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		// DEFAULTCAPTURE
		customname := values[2]
		if values[3] != "" {
			customname = values[3]
		}
		//search for replacements
		replace := fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[DEFAULTCAPTURE][values[2]])
		//build the new regex
		newPatternDF = strings.Replace(newPatternDF, values[0], replace, -1)

		// NC
		customname = values[2]
		if values[3] != "" {
			customname = values[3]
			//search for replacement
			replace := fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[NAMEDCAPTURE][values[2]])
			//build the new regex
			newPatternNC = strings.Replace(newPatternNC, values[0], replace, -1)
		} else {
			//search the replacement
			replace := finalPatterns[NAMEDCAPTURE][values[2]]
			//build the new regex
			newPatternNC = strings.Replace(newPatternNC, values[0], replace, -1)
		}

	}

	return newPatternDF, newPatternNC
}

func (g *Grok) Patterns() map[Option]map[string]string {

	return g.patterns
}
