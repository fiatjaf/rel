package main

import (
	"strings"

	"github.com/chzyer/readline"
	"github.com/renstrom/fuzzysearch/fuzzy"
)

func nodeAutocompleter(s *state) readline.AutoCompleter {
	haystack := make([]string, len(s.Nodes))
	i := 0
	for _, n := range s.Nodes {
		haystack[i] = n.repr()
		i++
	}

	return FuzzyCompleter{haystack}
}

type FuzzyCompleter struct {
	haystack []string
}

func (fz FuzzyCompleter) Do(line []rune, pos int) ([][]rune, int) {
	needle := string(line)
	results := fuzzy.Find(needle, fz.haystack)

	out := make([][]rune, len(results))
	for i, result := range results {
		out[i] = []rune(result)
	}

	return out, 0
}

func autocompleteNodes(s *state, prompt string) (*Node, error) {
	completer := nodeAutocompleter(s)

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m" + prompt + "\033[0m ",
		HistoryFile:       "/tmp/readline.tmp",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}

	for {
		v, err := l.Readline()
		if err != nil {
			return nil, err
		}

		v = strings.Trim(v, " ")
		parts := strings.Split(v, "(")
		if len(parts) != 2 {
			continue
		}
		id := strings.TrimRight(parts[1], ")")

		if n, found := s.Nodes[id]; found {
			return n, nil
		}
	}
}
