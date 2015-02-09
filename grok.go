package grok

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// var patterns = map[string]string{}

type Grok struct {
	compiledPattern *regexp.Regexp
	patterns        map[string]string
}

// New returns a Grok struct
func New() *Grok {
	o := new(Grok)
	o.patterns = map[string]string{}
	return o
}

// AddPattern add a named pattern to grok
func (g *Grok) AddPattern(name string, pattern string) {
	g.patterns[name] = pattern
}

// Patterns returns all loaded patterns
func (g *Grok) Patterns() map[string]string {
	return g.patterns
}

// Compile
func (g *Grok) Compile(pattern string) error {

	//TODO : Denormalize pattern
	//   "%{DAY}" => (?P<DAY>(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?))
	//   "%{DAY:foo} => (?P<foo>(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?))
	//   "%{DAY:foo}" .* => (?P<foo>(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?) .*)

	//search for %{...:...}
	r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
	newPattern := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		customname := names[0]
		if len(names) != 1 {
			customname = names[1]
		}
		//rechercher les replacements
		if ok := g.Patterns()[names[0]]; ok == "" {
			return fmt.Errorf("ERROR no pattern found for %%{%s}", names[0])
		}
		replace := fmt.Sprintf("(?P<%s>%s)", customname, g.Patterns()[names[0]])
		//construire la nouvelle regexp
		newPattern = strings.Replace(newPattern, values[0], replace, -1)
	}

	patternCompiled, err := regexp.Compile(newPattern)
	if err != nil {
		return err
	}

	g.compiledPattern = patternCompiled

	return nil
}

func (g *Grok) Match(text string) bool {
	if g.compiledPattern == nil {
		return false
	}

	if m := g.compiledPattern.MatchString(text); !m {
		return false
	}

	return true
}

func (g *Grok) Captures(text string) (map[string]string, error) {
	captures := make(map[string]string)
	if g.compiledPattern == nil {
		return captures, fmt.Errorf("missing compiled regexp")
	}

	match := g.compiledPattern.FindStringSubmatch(text)
	for i, name := range g.compiledPattern.SubexpNames() {
		captures[name] = match[i]
	}

	return captures, nil
}

func (g *Grok) AddPatternsFromFile(path string) error {
	inFile, _ := os.Open(path)
	defer inFile.Close()

	reader := bufio.NewReader(inFile)
	scanner := bufio.NewScanner(reader)

	var patternDependancies = graph{}
	var fileContent = map[string]string{}

	for scanner.Scan() {
		l := scanner.Text()
		if len(l) > 0 { // line has text
			if l[0] != '#' { // line does not start with #
				//fmt.Println(l)

				names := strings.Split(l, " ")
				// names[0] = key
				// names[1] = pattern
				fileContent[names[0]] = names[1]

				r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
				keys := []string{}
				for _, key := range r.FindAllStringSubmatch(names[1], -1) {
					rawKey := strings.Split(key[1], ":")
					keys = append(keys, rawKey[0])
				}
				//fmt.Printf("%s => %s\n", names[0], keys)
				patternDependancies[names[0]] = keys
			}
		}
	}

	order, _ := topSortDFS(patternDependancies)
	order = reverseList(order)

	var denormalizedPattern = map[string]string{}
	for _, key := range order {
		//fmt.Printf("%s => %s\n", key, fileContent[key])
		denormalizedPattern[key] = oo(fileContent[key], denormalizedPattern)
		g.AddPattern(key, denormalizedPattern[key])
	}

	return nil
}

func oo(pattern string, finalPatterns map[string]string) string {
	r, _ := regexp.Compile(`%{(\w+:?\w+)}`)
	newPattern := pattern
	for _, values := range r.FindAllStringSubmatch(pattern, -1) {
		names := strings.Split(values[1], ":")

		customname := names[0]
		if len(names) != 1 {
			customname = names[1]
		}
		//rechercher les replacements
		replace := fmt.Sprintf("(?P<%s>%s)", customname, finalPatterns[names[0]])
		//log.Printf("replace %s by %s", values[0], replace)
		//construire la nouvelle regexp
		newPattern = strings.Replace(newPattern, values[0], replace, -1)

	}
	return newPattern
}

func (g *Grok) Discover(text string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

type Pile struct {
	Patterns     map[string]string
	PatternFiles []string
	Groks        []*Grok
}

func NewPile() *Pile {
	pile := new(Pile)
	pile.Patterns = make(map[string]string)
	pile.PatternFiles = make([]string, 0)
	pile.Groks = make([]*Grok, 0)

	return pile
}

func (pile *Pile) AddPattern(name, str string) error {
	return fmt.Errorf("Not Implemented")
}

func (pile *Pile) Compile(pattern string) error {
	return fmt.Errorf("Not Implemented")
}

func (pile *Pile) AddPatternsFromFile(path string) {
	pile.PatternFiles = append(pile.PatternFiles, path)
}

func (pile *Pile) Match(str string) *Grok {
	return nil
}
