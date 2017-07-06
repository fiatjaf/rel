package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Songmu/prompter"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func main() {
	app := cli.NewApp()

	app.Name = "rels"
	app.Description = "manage entities and relationships between them in flat files."
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json",
			Usage: "get the output in JSON whenever possible",
		},
	}

	s := &state{
		Nodes: make(map[string]*Node),
		Rels:  make(map[string]*Rel),
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

			n := Node{
				path:  path,
				state: s,
			}
			err = yaml.Unmarshal(contents, &n)
			if err != nil {
				log.Print("couldn't parse yaml from file: ", path, ". ", err)
				return nil
			}

			s.Nodes[n.Id] = &n
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
							r := Rel{
								Kind:     kind.(string),
								Directed: spec != 0,
							}

							if spec > 0 {
								r.From = s.Nodes[this]
								r.To = s.Nodes[id.(string)]
							} else {
								r.To = s.Nodes[this]
								r.From = s.Nodes[id.(string)]
							}

							if _, exists := s.Rels[r.key()]; !exists {
								s.Rels[r.key()] = &r
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
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				if c.GlobalBool("json") {
					return json.NewEncoder(os.Stdout).Encode(s.Nodes)
				}

				for _, n := range s.Nodes {
					fmt.Println(n.repr())
				}
				return nil
			},
		},
		{
			Name:  "links",
			Usage: "list all relationships",
			Action: func(c *cli.Context) error {
				if c.GlobalBool("json") {
					return json.NewEncoder(os.Stdout).Encode(s.Rels)
				}

				for _, r := range s.Rels {
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

				n := addNode(s, name)
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

				r := Rel{
					Directed: !c.Bool("neutral"),
					Kind:     args[0],
				}

				if n, err := autocompleteNodes(s, "from:"); err != nil {
					return err
				} else {
					r.From = n
				}

				if n, err := autocompleteNodes(s, "to: "); err != nil {
					return err
				} else {
					r.To = n
				}

				s.Rels[r.key()] = &r

				r.From.write()
				r.To.write()

				return nil
			},
		},
		{
			Name:  "unlink",
			Usage: "remove a link",
			Action: func(c *cli.Context) error {
				if r, err := autocompleteRels(s, "link: "); err != nil {
					return err
				} else {
					delete(s.Rels, r.key())
					r.From.write()
					r.To.write()
					return nil
				}
			},
		},
		{
			Name:  "print",
			Usage: "print the contents of a node file",
			Action: func(c *cli.Context) error {
				if n, err := autocompleteNodes(s, "name: "); err != nil {
					return err
				} else {
					contents, err := ioutil.ReadFile(n.path)
					if err != nil {
						return err
					}
					_, err = os.Stdout.Write(contents)
					return err
				}
			},
		},
		{
			Name:  "edit",
			Usage: "open a file for edit by node name.",
			Action: func(c *cli.Context) error {
				if n, err := autocompleteNodes(s, "name: "); err != nil {
					return err
				} else {
					cmd := exec.Command("edit", n.path)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Start(); err != nil {
						return err
					}
					if err := cmd.Wait(); err != nil {
						return err
					}
					return nil
				}
			},
		},
		{
			Name:  "dot",
			Usage: "generate a dot string of the graph",
			Action: func(c *cli.Context) error {
				return dot.Execute(os.Stdout, s)
			},
		},
	}

	app.Run(os.Args)
}

var dot = template.Must(template.New("dot").Parse(`
digraph main {
  {{ range .Nodes }}
  n{{ .Id }} [label="{{ .Name }}"]; {{ end }}

  {{ range .Rels }}
  n{{ .From.Id }}->n{{ .To.Id }} [
    label="{{ .Kind }}",
    dir={{ if .Directed }}forward{{ else }}none{{ end }}
  ]; {{ end }}
}
`))
