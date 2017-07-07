package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/Songmu/prompter"
	"github.com/fiatjaf/cuid"
	"gopkg.in/yaml.v2"
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

	Id    string                 `json:"id"`
	Name  string                 `json:"name"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`

	Timestamp int64 `json:"timestamp"`
}

func (n Node) MarshalYAML() (interface{}, error) {
	raw := map[string]interface{}{
		"name":      n.Name,
		"id":        n.Id,
		"timestamp": n.Timestamp,
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
		} else if r.To.Id == n.Id {
			other = r.From.Id
			out = false
		} else {
			// this node is not participating in current rel
			continue
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
	if tm, ok := raw["timestamp"].(int64); ok {
		n.Timestamp = tm
	}
	if n.Name == "" || n.Id == "" {
		return fmt.Errorf("invalid file, missing name and/or id.")
	}

	n.Attrs = make(map[string]interface{})

	for k, v := range raw {
		if k != "name" && k != "id" &&
			k != "outgoing" && k != "incoming" && k != "neutral" &&
			k != "timestamp" {

			n.Attrs[k] = v
		}
	}

	return nil
}

func (n Node) repr() string {
	return n.Name + " (" + n.Id + ")"
}

func addNode(s *state, name string) *Node {
	id := cuid.Slug()
	for _, n := range s.Nodes {
		if n.Name == name {
			dup := prompter.YN(
				"There's already a node named '"+name+"', create a duplicate?",
				false,
			)
			if dup {
				break
			} else {
				return nil
			}
		}
	}

	n := &Node{
		path:      path.Join(s.here, id+".yaml"),
		state:     s,
		Id:        id,
		Name:      name,
		Timestamp: time.Now().Unix(),
	}
	s.Nodes[id] = n

	return n
}

func (n Node) write() error {
	contents, err := yaml.Marshal(n)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(n.path, contents, 0777)
}

type Rel struct {
	Kind     string `json:"kind"`
	Directed bool   `json:"directed"`
	From     *Node  `json:"from"`
	To       *Node  `json:"to"`
}

func (r Rel) key() string {
	var prekey string
	if r.Directed {
		prekey = r.From.Id + "-" + r.Kind + ">" + r.To.Id
	} else {
		// alphabetic
		if r.From.Id < r.To.Id {
			prekey = r.From.Id + "-" + r.Kind + "-" + r.To.Id
		} else {
			prekey = r.To.Id + "-" + r.Kind + "-" + r.From.Id
		}
	}

	hasher := md5.New()
	hasher.Write([]byte(prekey))
	return hex.EncodeToString(hasher.Sum(nil))[:5]
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
