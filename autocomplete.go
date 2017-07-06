package main

import (
	"strings"

	"github.com/chzyer/readline"
)

func nodeAutocompleter(s *state) *readline.PrefixCompleter {
	pc := readline.PrefixCompleter{}

	for _, n := range s.Nodes {
		pc.Children = append(
			pc.Children,
			readline.PcItem(n.repr()),
		)
	}

	return &pc
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

func relAutocompleter(s *state) *readline.PrefixCompleter {
	pc := readline.PrefixCompleter{}

	for _, r := range s.Rels {
		pc.Children = append(
			pc.Children,
			readline.PcItem(r.repr()+" ("+r.key()+")"),
		)
	}

	return &pc
}

func autocompleteRels(s *state, prompt string) (*Rel, error) {
	completer := relAutocompleter(s)

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
		key := strings.TrimRight(parts[1], ")")

		if k, found := s.Rels[key]; found {
			return k, nil
		}
	}
}
