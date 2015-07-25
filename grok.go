package grok

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// A Config structure is used to configure a Grok parser.
type Config struct {
	NamedCapturesOnly   bool
	SkipDefaultPatterns bool
	Patterns            map[string]string
}

// Grok object us used to load patterns and deconstruct strings using those
// patterns.
type Grok struct {
	config           *Config
	compiledPatterns map[string]*gRegexp
	patterns         map[string]*gPattern
	serviceMu        sync.Mutex
}

type gPattern struct {
	expression string
	typeInfo   semanticTypes
}

type gRegexp struct {
	regexp   *regexp.Regexp
	typeInfo semanticTypes
}

type semanticTypes map[string]string

// New returns a Grok object.
func New() *Grok {
	return NewWithConfig(&Config{})
}

// NewWithConfig returns a Grok object that is configured to behave according
// to the supplied Config structure.
func NewWithConfig(config *Config) *Grok {
	g := &Grok{config: config, compiledPatterns: map[string]*gRegexp{}}

	if config.Patterns != nil {
		g.AddPatternsFromMap(config.Patterns)
	}

	if g.patterns == nil {
		g.patterns = make(map[string]*gPattern)
	}

	if !config.SkipDefaultPatterns {
		g.AddPatternsFromMap(patterns)
	}

	return g
}

// Patterns return a map of the loaded patterns.
func (g *Grok) Patterns() map[string]*gPattern {
	return g.patterns
}

// AddPattern adds a new pattern to the list of loaded patterns.
func (g *Grok) AddPattern(name, pattern string) error {
	dnPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	if err != nil {
		return err
	}

	g.patterns[name] = &gPattern{expression: dnPattern, typeInfo: ti}
	return nil
}

// AddPatternsFromMap adds new patterns from the specified map to the list of
// loaded patterns.
func (g *Grok) AddPatternsFromMap(m map[string]string) error {
	re, _ := regexp.Compile(`%{(\w+):?(\w+)?}`)

	patternDeps := graph{}
	for k, v := range m {
		keys := []string{}
		for _, key := range re.FindAllStringSubmatch(v, -1) {
			keys = append(keys, key[1])
		}
		patternDeps[k] = keys
	}

	order, _ := sortGraph(patternDeps)

	for _, key := range reverseList(order) {
		g.AddPattern(key, m[key])
	}

	return nil
}

// AddPatternsFromPath adds new patterns from the files in the specified
// directory to the list of loaded patterns.
func (g *Grok) AddPatternsFromPath(path string) error {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path = path + "/*"
		}
	} else {
		return fmt.Errorf("invalid path : %s", path)
	}

	// only one error can be raised, when pattern is malformed
	// pattern is hard-coded "/*" so we ignore err
	files, _ := filepath.Glob(path)

	var filePatterns = map[string]string{}
	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				filePatterns[names[0]] = names[1]
			}
		}

		file.Close()
	}

	return g.AddPatternsFromMap(filePatterns)
}

// Match returns true if the specified text matches the pattern.
func (g *Grok) Match(pattern, text string) (bool, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return false, err
	}

	if ok := gr.regexp.MatchString(text); !ok {
		return false, nil
	}

	return true, nil
}

// Parse the specified text and return a map with the results.
func (g *Grok) Parse(pattern, text string) (map[string]string, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}

	captures := make(map[string]string)
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				captures[name] = match[i]
			}
		}
	}

	return captures, nil
}

// Parse returns a inteface{} map with captured fields based on provided pattern over the text
func (g *Grok) ParseTyped(pattern string, text string) (map[string]interface{}, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}
	match := gr.regexp.FindStringSubmatch(text)
	captures := make(map[string]interface{})
	if len(match) > 0 {
		for i, segmentName := range gr.regexp.SubexpNames() {
			if len(segmentName) != 0 {
				var value, err interface{}
				segmentType := gr.typeInfo[segmentName]
				switch segmentType {
				case "":
					value, err = match[i], nil
				case "int":
					value, err = strconv.ParseFloat(match[i], 64)
					value = int(value.(float64))
				case "float":
					value, err = strconv.ParseFloat(match[i], 64)
				default:
					return nil, fmt.Errorf("ERROR the value %s cannot be converted to %s", match[i], segmentType)
				}
				if err == nil {
					captures[segmentName] = value
				}
			}

		}
	}

	return captures, nil
}

// ParseToMultiMap parses the specified text and returns a map with the
// results. Values are stored in an string slice, so values from captures with
// the same name don't get overridden.
func (g *Grok) ParseToMultiMap(pattern, text string) (map[string][]string, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}

	captures := make(map[string][]string)
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				captures[name] = append(captures[name], match[i])
			}
		}
	}

	return captures, nil
}

func (g *Grok) cache(pattern string, cr *gRegexp) {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()
	g.compiledPatterns[pattern] = cr
}

func (g *Grok) cacheExists(pattern string) bool {
	g.serviceMu.Lock()
	defer g.serviceMu.Unlock()

	if _, ok := g.compiledPatterns[pattern]; ok {
		return true
	}

	return false
}

func (g *Grok) compile(pattern string) (*gRegexp, error) {
	if g.cacheExists(pattern) {
		return g.compiledPatterns[pattern], nil
	}

	newPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	if err != nil {
		return nil, err
	}

	compiledRegex, err := regexp.Compile(newPattern)
	if err != nil {
		return nil, err
	}
	gr := &gRegexp{regexp: compiledRegex, typeInfo: ti}
	g.cache(pattern, gr)
	return gr, nil
}

func (g *Grok) denormalizePattern(pattern string, storedPatterns map[string]*gPattern) (string, semanticTypes, error) {
	r, _ := regexp.Compile(`%{(\w+:?\w+:?\w+)}`)
	ti := semanticTypes{}
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		syntax, semantic := names[0], names[0]
		if len(names) > 1 {
			semantic = names[1]
		}

		// Add type cast information only if type set, and not string
		if len(names) == 3 {
			if names[2] != "string" {
				ti[semantic] = names[2]
			}
		}

		storedPattern, ok := storedPatterns[syntax]
		if !ok {
			return "", ti, fmt.Errorf("no pattern found for %%{%s}", syntax)
		}

		var replace string
		if !g.config.NamedCapturesOnly || (g.config.NamedCapturesOnly && len(names) > 1) {
			replace = fmt.Sprintf("(?P<%s>%s)", semantic, storedPattern.expression)
		} else {
			replace = "(" + storedPattern.expression + ")"
		}

		//Merge type Informations
		for k, v := range storedPattern.typeInfo {
			//Lastest type information is the one to keep in memory
			if _, ok := ti[k]; !ok {
				ti[k] = v
			}
		}

		pattern = strings.Replace(pattern, values[0], replace, -1)
	}

	return pattern, ti, nil
}
