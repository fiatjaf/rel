package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type state struct {
	here      string
	nodesbyid map[string]*node
	relsbykey map[string]*rel
	schema    schema
}

type node struct {
	path  string
	state *state

	id    string
	name  string
	attrs map[string]interface{}
}

func (n node) MarshalYAML() (interface{}, error) {
	raw := map[string]interface{}{
		"name": n.name,
		"id":   n.id,
	}

	for k, v := range n.attrs {
		raw[k] = v
	}

	outgoing := map[string][]string{}
	incoming := map[string][]string{}
	neutral := map[string][]string{}

	for _, r := range n.state.relsbykey {
		var other string
		var out bool
		if r.from.id == n.id {
			other = r.to.id
			out = true
		} else {
			other = r.from.id
			out = false
		}

		if r.directed {
			if out {
				outgoing[r.kind] = append(outgoing[r.kind], other)
			} else {
				incoming[r.kind] = append(incoming[r.kind], other)
			}
		} else {
			neutral[r.kind] = append(neutral[r.kind], other)
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

func (n *node) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if name, ok := raw["name"].(string); ok {
		n.name = name
	}
	if id, ok := raw["id"].(string); ok {
		n.id = id
	}
	if n.name == "" || n.id == "" {
	}

	n.attrs = make(map[string]interface{})

	for k, v := range raw {
		if k != "name" && k != "id" && k != "outgoing" && k != "incoming" && k != "neutral" {
			n.attrs[k] = v
		}
	}

	return nil
}

func (n node) repr() string {
	return n.name + " (" + n.id + ")"
}

func (n node) write() error {
	contents, err := yaml.Marshal(n)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(n.path, contents, 0777)
}

type rel struct {
	kind     string
	directed bool
	from     *node
	to       *node
}

func (r rel) key() string {
	if r.directed {
		return r.from.id + "-" + r.kind + ">" + r.to.id
	} else {
		// alphabetic
		if r.from.id < r.to.id {
			return r.from.id + "-" + r.kind + "-" + r.to.id
		} else {
			return r.to.id + "-" + r.kind + "-" + r.from.id
		}

	}
}

func (r rel) repr() string {
	if r.directed {
		return "[" + r.from.name + "] ={" + r.kind + "}=> [" + r.to.name + "]"
	} else {
		// alphabetic
		if r.from.name < r.to.name {
			return "[" + r.from.name + "] ={" + r.kind + "}= [" + r.to.name + "]"
		} else {
			return "[" + r.to.name + "] ={" + r.kind + "}= [" + r.from.name + "]"
		}

	}
}

type schema struct {
}
