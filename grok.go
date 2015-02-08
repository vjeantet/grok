package grok

import "fmt"

type Grok struct {
}

func New() *Grok {
	return new(Grok)
}

func (g *Grok) Free() {

}

func (g *Grok) Compile(pattern string) error {
	return fmt.Errorf("Not Implemented")
}

func (g *Grok) AddPattern(name string, pattern string) {

}

func (g *Grok) AddPatternsFromFile(path string) error {
	return fmt.Errorf("Not Implemented")
}

func (g *Grok) Discover(text string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

func (g *Grok) Match(text string) (*Match, error) {
	return new(Match), fmt.Errorf("Not Implemented")
}

type Match struct {
}

func (m *Match) Captures() (map[string][]string, error) {
	captures := make(map[string][]string)
	return captures, fmt.Errorf("Not Implemented")
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

func (pile *Pile) Free() {

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

func (pile *Pile) Match(str string) (*Grok, *Match) {
	return nil, nil
}
