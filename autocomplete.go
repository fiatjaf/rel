package main

import (
	"fmt"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/chzyer/readline"
)

func nodeAutocompleter(s *state) readline.AutoCompleter {
	words := make([]string, len(s.Nodes))
	i := 0
	for _, n := range s.Nodes {
		words[i] = n.repr() + ", "
		i++
	}

	return MultiCompleter{words}
}

type MultiCompleter struct {
	Words []string
}

func (mn MultiCompleter) runeWords() [][]rune {
	rw := make([][]rune, len(mn.Words))
	for i, w := range mn.Words {
		rw[i] = []rune(w)
	}
	return rw
}

func (mn MultiCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if len(line) == 0 {
		return mn.runeWords(), 0
	}

	used := map[string]bool{}
	var start int
	for i, r := range line {
		if r == ',' {
			used[strings.Trim(string(line[start:i]), " ")] = true
			start = i + 1
		}
		if i == pos {
			break
		}
	}

	current := string(line[start:])
	nextcomma := strings.Index(current, ",")
	if nextcomma != -1 {
		current = current[:nextcomma]
	}
	current = strings.Trim(current, " ")
	nchars := len(current)

	var matches [][]rune
	for _, word := range mn.Words {
		if _, isused := used[word]; isused {
			continue
		}
		if strings.HasPrefix(word, current) {
			matches = append(matches, []rune(word)[nchars:])
		}
	}

	return matches, nchars
}

func autocompleteNodes(s *state, prompt string) ([]*Node, error) {
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

		items := strings.Split(v, ",")
		var result []*Node

		for _, item := range items {
			item = strings.Trim(item, " )")
			if item == "" {
				continue
			}

			parts := strings.Split(item, "(")
			if len(parts) == 2 {
				id := strings.TrimRight(parts[1], "), ")
				if n, found := s.Nodes[id]; found {
					result = append(result, n)
				}
			} else if len(parts) == 1 {
				// attempt to create a new node
				if prompter.YN("Create a node named '"+parts[0]+"'?", false) {
					if n := addNode(s, parts[0]); n != nil {
						result = append(result, n)
					}
				}
			}
		}

		if len(result) == 0 {
			return nil, fmt.Errorf("couldn't find or create node")
		}

		return result, nil
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
