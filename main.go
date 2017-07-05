package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/fiatjaf/cuid"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func main() {
	app := cli.NewApp()

	app.Name = "rels"
	app.Description = "manage entities and relationships between them in flat files."
	app.Version = "0.0.1"

	s := &state{}

	app.Before = func(c *cli.Context) error {
		here, _ := os.Getwd()

		s.here = here

		filepath.Walk(s.here, func(path string, f os.FileInfo, err error) error {
			if f.Name()[0] == '.' {
				if f.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
				return nil
			}

			contents, err := ioutil.ReadFile(path)
			if err != nil {
				log.Print("couldn't read file: ", path, ". ", err)
				return nil
			}

			n := node{
				path:  path,
				state: s,
			}
			err = yaml.Unmarshal(contents, &n)
			if err != nil {
				log.Print("couldn't parse yaml from file: ", path, ". ", err)
				return nil
			}

			if n.id == "" {
				n.id = strings.Split(filepath.Base(path), ".")[0]
			}

			s.nodes = append(s.nodes, &n)
			return nil
		})

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "nodes",
			Usage: "list all nodes",
			Action: func(c *cli.Context) error {
				for _, node := range s.nodes {
					fmt.Println(node.name)
				}
				return nil
			},
		},
		{
			Name:  "rels",
			Usage: "list all relationships",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
		{
			Name:  "add",
			Usage: "add a node",
			Action: func(c *cli.Context) error {
				name := c.Args().First()
				if name == "" {
					name = prompter.Prompt("Node name", "")
				}

				id := cuid.Slug()

				for _, n := range s.nodes {
					if n.name == name {
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

				n := node{
					path:  path.Join(s.here, id+".yaml"),
					state: s,

					id:   id,
					name: name,
				}
				return n.write()
			},
		},
		{
			Name:  "link",
			Usage: "create a relationship between two nodes",
			Action: func(c *cli.Context) error {
				fmt.Println("completed task: ", c.Args().First())
				return nil
			},
		},
	}

	app.Run(os.Args)
}

type state struct {
	here   string
	nodes  []*node
	rels   []*rel
	schema schema
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

	for _, r := range n.state.rels {
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

	n.name = raw["name"].(string)
	n.id = raw["id"].(string)
	n.attrs = make(map[string]interface{})

	for k, v := range raw {
		if k != "name" && k != "id" && k != "outgoing" && k != "incoming" && k != "neutral" {
			n.attrs[k] = v
		}
	}

	return nil
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

type schema struct {
}
