package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

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
	app.EnableBashCompletion = true

	s := &state{
		nodesbyid: make(map[string]*node),
		relsbykey: make(map[string]*rel),
	}

	app.Before = func(c *cli.Context) error {
		here, _ := os.Getwd()

		s.here = here

		// read all nodes
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

			s.nodesbyid[n.id] = &n
			return nil
		})

		// read all relationships
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

			var raw map[string]interface{}
			if err := yaml.Unmarshal(contents, &raw); err != nil {
				return err
			}

			var this string
			if id, ok := raw["id"].(string); ok {
				this = id
			}

			for reltype, spec := range map[string]int{
				"outgoing": 1,
				"neutral":  0,
				"incoming": -1,
			} {
				if relationships, ok := raw[reltype].(map[interface{}]interface{}); ok {
					for kind, ids := range relationships {
						for _, id := range ids.([]interface{}) {
							r := rel{
								kind:     kind.(string),
								directed: spec != 0,
							}

							if spec > 0 {
								r.from = s.nodesbyid[this]
								r.to = s.nodesbyid[id.(string)]
							} else {
								r.to = s.nodesbyid[this]
								r.from = s.nodesbyid[id.(string)]
							}

							if _, exists := s.relsbykey[r.key()]; !exists {
								s.relsbykey[r.key()] = &r
							}
						}
					}
				}
			}

			return nil
		})

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "nodes",
			Usage: "list all nodes",
			Action: func(c *cli.Context) error {
				for _, n := range s.nodesbyid {
					fmt.Println(n.repr())
				}
				return nil
			},
		},
		{
			Name:  "rels",
			Usage: "list all relationships",
			Action: func(c *cli.Context) error {
				for _, r := range s.relsbykey {
					fmt.Println(r.repr() + "\t(" + r.key() + ")")
				}
				return nil
			},
		},
		{
			Name:  "add",
			Usage: "add a node",
			Action: func(c *cli.Context) error {
				name := c.Args().First()
				if name == "" {
					name = prompter.Prompt("name", "")
				}

				id := cuid.Slug()

				for _, n := range s.nodesbyid {
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
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "neutral",
					Usage: "Use this if this relationship is not directed.",
				},
			},
			Action: func(c *cli.Context) error {
				args := c.Args()
				arglen := len(args)
				if arglen != 1 {
					return fmt.Errorf("argument should be <kind>")
				}

				r := rel{
					directed: !c.Bool("neutral"),
					kind:     args[0],
				}

				if n, err := autocompleteNodes(s, "from:"); err != nil {
					return err
				} else {
					r.from = n
				}

				if n, err := autocompleteNodes(s, "to: "); err != nil {
					return err
				} else {
					r.to = n
				}

				s.relsbykey[r.key()] = &r

				r.from.write()
				r.to.write()

				return nil
			},
		},
	}

	app.Run(os.Args)
}
