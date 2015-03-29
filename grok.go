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

// Grok Type
type Grok struct {
	compiledPattern map[string]*regexp.Regexp
	patterns        map[string]string
	serviceMu       sync.Mutex
}

// New returns a Grok struct
func New() *Grok {
	o := new(Grok)
	o.patterns = patterns
	o.compiledPattern = map[string]*regexp.Regexp{}
	return o
}

// AddPattern add a named pattern to grok
func (g *Grok) AddPattern(name string, pattern string) {
	g.patterns[name] = pattern
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
	for i, name := range cr.SubexpNames() {
		if len(match) > 0 {
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
	for i, name := range cr.SubexpNames() {
		if len(match) > 0 {
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
	}

	var patternDependancies = graph{}
	var fileContent = map[string]string{}

	//List all files if path folder
	files, _ := filepath.Glob(path)
	for _, file := range files {
		inFile, _ := os.Open(file)

		reader := bufio.NewReader(inFile)
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 { // line has text
				if l[0] != '#' { // line does not start with #
					names := strings.SplitN(l, " ", 2)
					// names[0] = key
					// names[1] = pattern
					fileContent[names[0]] = names[1]

					r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
					keys := []string{}
					for _, key := range r.FindAllStringSubmatch(names[1], -1) {
						rawKey := strings.Split(key[1], ":")
						keys = append(keys, rawKey[0])
					}
					patternDependancies[names[0]] = keys
				}
			}
		}
		inFile.Close()
	}

	order, _ := sortGraph(patternDependancies)
	order = reverseList(order)

	var denormalizedPattern = map[string]string{}
	for _, key := range order {
		denormalizedPattern[key] = denormalizePattern(fileContent[key], denormalizedPattern)
		g.AddPattern(key, denormalizedPattern[key])
	}

	return nil
}

func denormalizePattern(pattern string, finalPatterns map[string]string) string {
	r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
	newPattern := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		customname := names[0]
		if len(names) != 1 {
			customname = names[1]
		}
		//search for replacements
		replace := fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[names[0]])

		//build the new regex
		newPattern = strings.Replace(newPattern, values[0], replace, -1)
	}
	return newPattern
}
