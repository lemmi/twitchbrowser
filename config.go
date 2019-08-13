package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/lemmi/closer"
	"github.com/pkg/errors"
)

var (
	twitchlexer = lexer.Must(lexer.Regexp(`` +
		`(?m)` +
		`(?s)` +
		`(\s+)` +
		`|\[` +
		`|\]` +
		`|(^#.*$)` +
		`|(?P<Ident>[a-zA-Z_\d\-]+)`,
	))
	parser = participle.MustBuild(&TwitchConfig{}, participle.Lexer(twitchlexer))
)

// TwitchConfig is the structure of the config file
type TwitchConfig struct {
	Entry   Entries   `@@`
	Section []Section `@@*`
}

// Section represents a section within a config file
type Section struct {
	Name  string  `"[" @Ident "]"`
	Entry Entries `@@`
}

// Entries contains all entries within a section
type Entries struct {
	Value []string `@Ident*`
}

func (t TwitchConfig) config() configFile {
	c := make(configFile)

	c[""] = t.Entry.Value

	for _, section := range t.Section {
		c[section.Name] = section.Entry.Value
	}

	return c
}
func (t TwitchConfig) String() string {
	var s strings.Builder
	s.WriteString(t.Entry.String())

	if s.Len() > 0 {
		s.WriteString("\n")
	}

	for _, sec := range t.Section {
		s.WriteString("\n")
		s.WriteString(sec.String())
	}

	return s.String()
}
func (s Section) String() string {
	var ret string
	es := s.Entry.String()

	if len(es) > 0 {
		ret = "\n"
	}
	return fmt.Sprintf("[%s]\n%s%s", s.Name, es, ret)
}
func (es Entries) String() string {
	return strings.Join(es.Value, "\n")
}

type configFile map[string][]string

func parseConfig(r io.Reader) (configFile, error) {
	var t TwitchConfig
	if err := parser.Parse(r, &t); err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}

	return t.config(), nil
}

func loadConfigFile() (configFile, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(user.HomeDir, ".config", "twitchbrowser", "favorites.conf")
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %q", path)
	}
	defer closer.Do(file)

	return parseConfig(file)
}
