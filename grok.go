package grok

type Grok struct {
}

func New() *Grok {
	return new(Grok)
}

func Free() {

}

func (g *Grok) Compile(pattern string) error {
	return nil
}

func (g *Grok) AddPattern(name string, pattern string) {

}

func (g *Grok) AddPatternsFromFile(path string) error {
	return nil
}

func (g *Grok) Discover(text string) (string, error) {
	return "", nil
}

func (g *Grok) Match(text string) (*Match, error) {
	return new(Match), nil
}

type Match struct {
}

func (m *Match) Captures() map[string][]string {
	captures := make(map[string][]string)
	return captures
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
	return nil
}

func (pile *Pile) Compile(pattern string) error {
	return nil
}

func (pile *Pile) AddPatternsFromFile(path string) {
	pile.PatternFiles = append(pile.PatternFiles, path)
}

func (pile *Pile) Match(str string) (*Grok, *Match) {
	return nil, nil
}
