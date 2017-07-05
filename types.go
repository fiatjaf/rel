package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type state struct {
	here   string
	Nodes  map[string]*Node
	Rels   map[string]*Rel
	schema schema
}

type Node struct {
	path  string
	state *state

	Id    string
	Name  string
	Attrs map[string]interface{}
}

func (n Node) MarshalYAML() (interface{}, error) {
	raw := map[string]interface{}{
		"name": n.Name,
		"id":   n.Id,
	}

	for k, v := range n.Attrs {
		raw[k] = v
	}

	outgoing := map[string][]string{}
	incoming := map[string][]string{}
	neutral := map[string][]string{}

	for _, r := range n.state.Rels {
		var other string
		var out bool
		if r.From.Id == n.Id {
			other = r.To.Id
			out = true
		} else {
			other = r.From.Id
			out = false
		}

		if r.Directed {
			if out {
				outgoing[r.Kind] = append(outgoing[r.Kind], other)
			} else {
				incoming[r.Kind] = append(incoming[r.Kind], other)
			}
		} else {
			neutral[r.Kind] = append(neutral[r.Kind], other)
		}
	}

	if len(outgoing) > 0 {
		raw["outgoing"] = outgoing
	}
	if len(incoming) > 0 {
		raw["incoming"] = incoming
	}
	if len(neutral) > 0 {
		raw["neutral"] = neutral
	}

	return raw, nil
}

func (n *Node) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if name, ok := raw["name"].(string); ok {
		n.Name = name
	}
	if id, ok := raw["id"].(string); ok {
		n.Id = id
	}
	if n.Name == "" || n.Id == "" {
	}

	n.Attrs = make(map[string]interface{})

	for k, v := range raw {
		if k != "name" && k != "id" && k != "outgoing" && k != "incoming" && k != "neutral" {
			n.Attrs[k] = v
		}
	}

	return nil
}

func (n Node) repr() string {
	return n.Name + " (" + n.Id + ")"
}

func (n Node) write() error {
	contents, err := yaml.Marshal(n)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(n.path, contents, 0777)
}

type Rel struct {
	Kind     string
	Directed bool
	From     *Node
	To       *Node
}

func (r Rel) key() string {
	if r.Directed {
		return r.From.Id + "-" + r.Kind + ">" + r.To.Id
	} else {
		// alphabetic
		if r.From.Id < r.To.Id {
			return r.From.Id + "-" + r.Kind + "-" + r.To.Id
		} else {
			return r.To.Id + "-" + r.Kind + "-" + r.From.Id
		}

	}
}

func (r Rel) repr() string {
	if r.Directed {
		return "[" + r.From.Name + "] ={" + r.Kind + "}=> [" + r.To.Name + "]"
	} else {
		// alphabetic
		if r.From.Name < r.To.Name {
			return "[" + r.From.Name + "] ={" + r.Kind + "}= [" + r.To.Name + "]"
		} else {
			return "[" + r.To.Name + "] ={" + r.Kind + "}= [" + r.From.Name + "]"
		}

	}
}

type schema struct {
}
