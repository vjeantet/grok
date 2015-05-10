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
	compiledPattern    map[string]*regexp.Regexp
	patterns           map[string]string
	serviceMu          sync.Mutex
	defaultCaptureMode Option
}

// New returns a Grok struct
//  Options availables (any order) :
//  * grok.NODEFAULTPATTERNS => Do not use compiled-in patterns
//  * grok.DEFAULTCAPTURE    => Parse and ParseToMulti will return all groups
//  * grok.NAMEDCAPTURE      => Parse and ParseToMulti will return only named captures
//
//  Exemple
//  g := grok.New()
//  g := grok.New(grok.NAMEDCAPTURE)
//  g := grok.New(grok.NAMEDCAPTURE, grok.NODEFAULTPATTERNS)
//  g := grok.New(grok.NODEFAULTPATTERNS, grok.DEFAULTCAPTURE)
func New(opt ...Option) *Grok {
	o := new(Grok)
	o.defaultCaptureMode = DEFAULTCAPTURE
	o.patterns = map[string]string{}
	o.compiledPattern = map[string]*regexp.Regexp{}

	var skipDefaultPatterns bool
	for _, v := range opt {
		if v == NODEFAULTPATTERNS {
			skipDefaultPatterns = true
		}
		if v == NAMEDCAPTURE {
			o.defaultCaptureMode = NAMEDCAPTURE
		}
	}

	if skipDefaultPatterns == false {
		if o.defaultCaptureMode == NAMEDCAPTURE {
			o.patterns = namedCapturePatterns
		}
		if o.defaultCaptureMode == DEFAULTCAPTURE {
			o.patterns = defaultCapturePatterns
		}
	}

	return o
}

// AddPattern add a pattern to grok
func (g *Grok) AddPattern(name string, pattern string) {
	regexPattern := denormalizePattern(pattern, g.patterns, g.defaultCaptureMode)
	g.patterns[name] = regexPattern
}

func (g *Grok) cache(pattern string, cr *regexp.Regexp) {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()
	g.compiledPattern[pattern] = cr
}

func (g *Grok) cacheExists(pattern string) bool {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()

	if _, ok := g.compiledPattern[pattern]; ok {
		return true
	}

	return false
}

func (g *Grok) compile(pattern string) (*regexp.Regexp, error) {
	if g.cacheExists(pattern) {
		return g.compiledPattern[pattern], nil
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
		if ok := g.patterns[names[0]]; ok == "" {
			return nil, fmt.Errorf("ERROR no pattern found for %%{%s}", names[0])
		}
		replace := fmt.Sprintf("(?P<%s>%s)", customname, g.patterns[names[0]])

		//build the new regexp
		newPattern = strings.Replace(newPattern, values[0], replace, -1)
	}
	patternCompiled, err := regexp.Compile(newPattern)

	if err != nil {
		return nil, err
	}
	g.cache(pattern, patternCompiled)
	return patternCompiled, nil

}

// Match returns true when text match the compileed pattern
func (g *Grok) Match(pattern, text string) (bool, error) {
	cr, err := g.compile(pattern)

	if err != nil {
		return false, err
	}

	if m := cr.MatchString(text); !m {
		return false, nil
	}

	return true, nil
}

// Parse returns a string map with captured string based on provided pattern over the text
func (g *Grok) Parse(pattern string, text string) (map[string]string, error) {
	cr, err := g.compile(pattern)
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
func (g *Grok) ParseToMultiMap(pattern string, text string) (map[string][]string, error) {
	multiCaptures := make(map[string][]string)
	cr, err := g.compile(pattern)
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

func denormalizePattern(pattern string, finalPatterns map[string]string, kindOfCapture Option) string {
	r, _ := regexp.Compile(`%{((\w+):?(\w+)?)}`)
	newPattern := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		var replace string
		if kindOfCapture == DEFAULTCAPTURE {
			customname := values[2]
			if values[3] != "" {
				customname = values[3]
			}
			//search for replacements
			replace = fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[values[2]])
		}
		if kindOfCapture == NAMEDCAPTURE {
			customname := values[2]
			if values[3] != "" {
				customname = values[3]
				//search for replacement
				replace = fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[values[2]])
			} else {
				//search the replacement
				replace = finalPatterns[values[2]]
			}
		}
		//build the new regex
		newPattern = strings.Replace(newPattern, values[0], replace, -1)
	}

	return newPattern
}

//Patterns returns loaded registered patterns
func (g *Grok) Patterns() map[string]string {

	return g.patterns
}
