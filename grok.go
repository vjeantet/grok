package grok

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	valid  = regexp.MustCompile(`^\w+([-.]\w+)*(:(([-.()\w]+)|(\[\w+\])+)(:(string|float|int))?)?$`)
	normal = regexp.MustCompile(`%{([\w-.]+(?::[\w-.()\[\]]+(?::[\w-.()]+)?)?)}`)
	nested = regexp.MustCompile(`\[(\w+)\]`)
)

// A Config structure is used to configure a Grok parser.
type Config struct {
	NamedCapturesOnly   bool
	SkipDefaultPatterns bool
	RemoveEmptyValues   bool
	PatternsDir         []string
	Patterns            map[string]string
}

// Grok object us used to load patterns and deconstruct strings using those
// patterns.
type Grok struct {
	rawPattern       map[string]string
	config           *Config
	aliases          map[string]string
	compiledPatterns map[string]*gRegexp
	patterns         map[string]*gPattern
	patternsGuard    *sync.RWMutex
	compiledGuard    *sync.RWMutex
	aliasesGuard     *sync.RWMutex
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
func New() (*Grok, error) {
	return NewWithConfig(&Config{})
}

// NewWithConfig returns a Grok object that is configured to behave according
// to the supplied Config structure.
func NewWithConfig(config *Config) (*Grok, error) {
	g := &Grok{
		config:           config,
		aliases:          map[string]string{},
		compiledPatterns: map[string]*gRegexp{},
		patterns:         map[string]*gPattern{},
		rawPattern:       map[string]string{},
		patternsGuard:    new(sync.RWMutex),
		compiledGuard:    new(sync.RWMutex),
		aliasesGuard:     new(sync.RWMutex),
	}

	if !config.SkipDefaultPatterns {
		err := g.AddPatternsFromMap(patterns)
		if err != nil {
			return nil, err
		}
	}

	if len(config.PatternsDir) > 0 {
		for _, path := range config.PatternsDir {
			err := g.AddPatternsFromPath(path)
			if err != nil {
				return nil, err
			}
		}

	}

	if err := g.AddPatternsFromMap(config.Patterns); err != nil {
		return nil, err
	}

	return g, nil
}

// AddPattern adds a new pattern to the list of loaded patterns.
func (g *Grok) addPattern(name, pattern string) error {
	dnPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	if err != nil {
		return err
	}

	g.patterns[name] = &gPattern{expression: dnPattern, typeInfo: ti}
	return nil
}

// AddPattern adds a named pattern to grok
func (g *Grok) AddPattern(name, pattern string) error {
	g.patternsGuard.Lock()
	defer g.patternsGuard.Unlock()

	g.rawPattern[name] = pattern
	return g.buildPatterns()
}

// AddPatternsFromMap loads a map of named patterns
func (g *Grok) AddPatternsFromMap(m map[string]string) error {
	g.patternsGuard.Lock()
	defer g.patternsGuard.Unlock()

	for name, pattern := range m {
		g.rawPattern[name] = pattern
	}
	return g.buildPatterns()
}

// AddPatternsFromMap adds new patterns from the specified map to the list of
// loaded patterns.
func (g *Grok) addPatternsFromMap(m map[string]string) error {
	patternDeps := graph{}
	for k, v := range m {
		var keys []string
		for _, key := range normal.FindAllStringSubmatch(v, -1) {
			if !valid.MatchString(key[1]) {
				return fmt.Errorf("invalid pattern %%{%s}", key[1])
			}
			names := strings.Split(key[1], ":")
			syntax := names[0]
			if g.patterns[syntax] == nil {
				if _, ok := m[syntax]; !ok {
					return fmt.Errorf("no pattern found for %%{%s}", syntax)
				}
			}
			keys = append(keys, syntax)
		}
		patternDeps[k] = keys
	}
	order, _ := sortGraph(patternDeps)
	for _, key := range reverseList(order) {
		err := g.addPattern(key, m[key])
		if err != nil {
			return fmt.Errorf("cannot add pattern %q: %v", key, err)
		}
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

		_ = file.Close()
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

// compiledParse parses the specified text and returns a map with the results.
func (g *Grok) compiledParse(gr *gRegexp, text string) (map[string]string, error) {
	captures := make(map[string]string, gr.regexp.NumSubexp())
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				if g.config.RemoveEmptyValues && match[i] == "" {
					continue
				}
				name = g.nameToAlias(name)
				captures[name] = match[i]
			}
		}
	}

	return captures, nil
}

// Parse the specified text and return a map with the results.
func (g *Grok) Parse(pattern, text string) (map[string]string, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}

	return g.compiledParse(gr, text)
}

// ParseTyped returns a interface{} map with typed captured fields based on provided pattern over the text.
// Is able to return nested map[string]interface{} maps when %{PATTERN:[nested][field]} syntax is used.
func (g *Grok) ParseTyped(pattern string, text string) (map[string]interface{}, error) {
	gr, err := g.compile(pattern)
	if err != nil {
		return nil, err
	}
	match := gr.regexp.FindStringSubmatch(text)
	captures := make(map[string]interface{}, gr.regexp.NumSubexp())
	if len(match) > 0 {
		for i, segmentName := range gr.regexp.SubexpNames() {
			if len(segmentName) != 0 {
				if g.config.RemoveEmptyValues == true && match[i] == "" {
					continue
				}
				name := g.nameToAlias(segmentName)
				nested_path := []string{}
				nested_names := nested.FindAllStringSubmatch(name, -1)

				if nested_names != nil {
					for _, element := range nested_names {
						nested_path = append(nested_path, element[1])
					}
				}

				if segmentType, ok := gr.typeInfo[name]; ok {
					switch segmentType {
					case "int":
						value, _ := strconv.Atoi(match[i])
						if len(nested_path) > 0 {
							addNested(captures, nested_path, value)
						} else {
							captures[name] = value
						}
					case "float":
						value, _ := strconv.ParseFloat(match[i], 64)
						if len(nested_path) > 0 {
							addNested(captures, nested_path, value)
						} else {
							captures[name] = value
						}
					default:
						return nil, fmt.Errorf("ERROR the value %s cannot be converted to %s", match[i], segmentType)
					}
				} else {
					if len(nested_path) > 0 {
						addNested(captures, nested_path, match[i])
					} else {
						captures[name] = match[i]
					}
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

	captures := make(map[string][]string, gr.regexp.NumSubexp())
	if match := gr.regexp.FindStringSubmatch(text); len(match) > 0 {
		for i, name := range gr.regexp.SubexpNames() {
			if name != "" {
				if g.config.RemoveEmptyValues == true && match[i] == "" {
					continue
				}
				name = g.nameToAlias(name)
				captures[name] = append(captures[name], match[i])
			}
		}
	}

	return captures, nil
}

func (g *Grok) buildPatterns() error {
	g.patterns = map[string]*gPattern{}
	return g.addPatternsFromMap(g.rawPattern)
}

func (g *Grok) compile(pattern string) (*gRegexp, error) {
	g.compiledGuard.RLock()
	gr, ok := g.compiledPatterns[pattern]
	g.compiledGuard.RUnlock()

	if ok {
		return gr, nil
	}

	g.patternsGuard.RLock()
	newPattern, ti, err := g.denormalizePattern(pattern, g.patterns)
	g.patternsGuard.RUnlock()
	if err != nil {
		return nil, err
	}

	compiledRegex, err := regexp.Compile(newPattern)
	if err != nil {
		return nil, err
	}
	gr = &gRegexp{regexp: compiledRegex, typeInfo: ti}

	g.compiledGuard.Lock()
	g.compiledPatterns[pattern] = gr
	g.compiledGuard.Unlock()

	return gr, nil
}

func (g *Grok) denormalizePattern(pattern string, storedPatterns map[string]*gPattern) (string, semanticTypes, error) {
	ti := semanticTypes{}
	matches := normal.FindAllStringSubmatchIndex(pattern, -1)
	if len(matches) == 0 {
		return pattern, ti, nil
	}

	var result strings.Builder
	result.Grow(len(pattern) * 2) // Pre-allocate with estimate
	lastEnd := 0

	for _, match := range matches {
		matchStart := match[0]
		matchEnd := match[1]
		submatchStart := match[2]
		submatchEnd := match[3]

		// Extract the matched pattern name (e.g., "WORD:field:int")
		patternName := pattern[submatchStart:submatchEnd]

		if !valid.MatchString(patternName) {
			return "", ti, fmt.Errorf("invalid pattern %%{%s}", patternName)
		}

		names := strings.Split(patternName, ":")
		syntax, semantic, alias := names[0], names[0], names[0]
		if len(names) > 1 {
			semantic = names[1]
			alias = g.aliasizePatternName(semantic)
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

		// Copy text before this match
		result.WriteString(pattern[lastEnd:matchStart])

		// Build replacement
		if !g.config.NamedCapturesOnly || (g.config.NamedCapturesOnly && len(names) > 1) {
			result.WriteString("(?P<")
			result.WriteString(alias)
			result.WriteString(">")
			result.WriteString(storedPattern.expression)
			result.WriteString(")")
		} else {
			result.WriteString("(")
			result.WriteString(storedPattern.expression)
			result.WriteString(")")
		}

		// Merge type Information
		for k, v := range storedPattern.typeInfo {
			// Latest type information is the one to keep in memory
			if _, ok := ti[k]; !ok {
				ti[k] = v
			}
		}

		lastEnd = matchEnd
	}

	// Copy remaining text after last match
	result.WriteString(pattern[lastEnd:])

	return result.String(), ti, nil
}

func (g *Grok) aliasizePatternName(name string) string {
	d := []byte(name)
	alias := fmt.Sprintf("h%x", md5.Sum(d))
	g.aliasesGuard.Lock()
	g.aliases[alias] = name
	g.aliasesGuard.Unlock()
	return alias
}

func (g *Grok) nameToAlias(name string) string {
	g.aliasesGuard.RLock()
	alias, ok := g.aliases[name]
	g.aliasesGuard.RUnlock()
	if ok {
		return alias
	}
	return name
}

// ParseStream will match the given pattern on a line by line basis from the reader
// and apply the results to the process function
func (g *Grok) ParseStream(reader *bufio.Reader, pattern string, process func(map[string]string) error) error {
	gr, err := g.compile(pattern)
	if err != nil {
		return err
	}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		values, err := g.compiledParse(gr, line)
		if err != nil {
			return err
		}
		if err = process(values); err != nil {
			return err
		}
	}
}

// adds a variable to a string keyed map going as deep as needed
func addNested(n map[string]interface{}, path []string, value interface{}) error {
	//pop path element => current element
	element, path := path[0], path[1:]

	//if this is the leaf element of the path
	//just add it to the map
	if len(path) == 0 {
		n[element] = value
		return nil
	}

	var childmap map[string]interface{}
	var ismap bool

	//check whether the current element already exists and is a map
	child, exists := n[element]
	if exists {
		childmap, ismap = child.(map[string]interface{})
		if !ismap { //in case the current element does exist but is not map it's not possible to walk down the path
			return fmt.Errorf("Nesting under an already used key")
		}
	} else {
		//in case the current element does NOT exist make a map
		childmap = make(map[string]interface{})
		n[element] = childmap
	}

	//and finally walk down the path recursively
	return addNested(childmap, path, value)
}
