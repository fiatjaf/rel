package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Songmu/prompter"
	"github.com/thomasheller/gtf/gtft"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func main() {
	app := cli.NewApp()

	app.Name = "rels"
	app.Description = "manage entities and relationships between them in flat files."
	app.Version = "0.2.0"
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
				list := nodeMapToList(s.Nodes)
				sort.Sort(byRepr(list))
				if c.GlobalBool("json") {
					return json.NewEncoder(os.Stdout).Encode(list)
				}

				for _, in := range list {
					n := in.(*Node)
					fmt.Println(n.repr())
				}
				return nil
			},
		},
		{
			Name:  "links",
			Usage: "list all relationships",
			Action: func(c *cli.Context) error {
				list := relMapToList(s.Rels)
				sort.Sort(byRepr(list))

				if c.GlobalBool("json") {
					return json.NewEncoder(os.Stdout).Encode(list)
				}

				for _, ir := range list {
					r := ir.(*Rel)
					fmt.Println(r.repr() + " (" + r.key() + ")")
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
				if err := n.write(); err != nil {
					return err
				}

				fmt.Println("created " + n.repr())
				return nil
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

				fromnodes, err := autocompleteNodes(s, "from:")
				if err != nil {
					return err
				}

				tonodes, err := autocompleteNodes(s, "to: ")
				if err != nil {
					return err
				}

				for _, fromnode := range fromnodes {
					for _, tonode := range tonodes {
						r := Rel{
							Directed: !c.Bool("neutral"),
							Kind:     args[0],
							From:     fromnode,
							To:       tonode,
						}

						s.Rels[r.key()] = &r
						r.From.write()
						r.To.write()

						fmt.Println("created " + r.repr())
					}
				}

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
				if nodes, err := autocompleteNodes(s, "name: "); err != nil {
					return err
				} else {
					for _, n := range nodes {
						fmt.Println(n.path)
						contents, err := ioutil.ReadFile(n.path)
						if err != nil {
							return err
						}
						if _, err = os.Stdout.Write(contents); err != nil {
							log.Print(err)
						}
						fmt.Println("")
					}
					return nil
				}
			},
		},
		{
			Name:  "edit",
			Usage: "open a file for edit by node name.",
			Action: func(c *cli.Context) error {
				if nodes, err := autocompleteNodes(s, "name: "); err != nil {
					return err
				} else {
					for _, n := range nodes {
						editor := os.Getenv("EDITOR")
						if editor == "" {
							editor = "edit"
						}
						cmd := exec.Command(editor, n.path)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr

						if err := cmd.Start(); err != nil {
							return err
						}
						if err := cmd.Wait(); err != nil {
							return err
						}
					}
					return nil
				}
			},
		},
		{
			Name:  "dot",
			Usage: "generate a dot string of the graph",
			Action: func(c *cli.Context) error {
				dot := template.Must(template.New("~").Parse(dotTemplate))
				return dot.Execute(os.Stdout, s)
			},
		},
		{
			Name:  "template",
			Usage: "run a Go template against the data of the graph",
			Description: `For any kind of rendering, like HTML, CSV or even DOT (if you want something that goes beyond the really basic template that is provided by 'rel dot') this is the command you're looking for. The template is any text file using the syntax specified by https://golang.org/pkg/text/template/ and the extra helper functions given by https://github.com/leekchan/gtf, to which will be passed the variable '.Nodes' -- a list of all the nodes, in the graph, and '.Rels', a list of all links in the graph. For an example look at the template that is being used to render the DOT output given by 'rel dot'.

   If you want to exclude some nodes from rendering you can either pass a list of the nodes as an argument to --exclude comma-separated or pipe a list of nodes newline-separated to the STDIN of the program:

    cat my-exclusion-list | rel template --template something.dot --exclude 'Thing that was excluded (83h4s2l), Excluded thing (si834sp)'

  The items in both exclusion list must be either the raw node id or some name followed by the node id inside parenthesis. This makes easy to filter by the following method:

    rel nodes > my-exclusion-list
    vim my-exclusion-list (filter it manually)
    cat my-exclusion list | rel template --template my-template.dot
            `,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "template",
					Usage: "template file to be used.",
				},
				cli.StringFlag{
					Name:  "exclude",
					Usage: "list of nodes to exclude.",
				},
			},
			Action: func(c *cli.Context) error {
				file := c.String("template")
				tmpl, err := template.New(filepath.Base(file)).Funcs(gtft.GtfFuncMap).ParseFiles(file)
				if err != nil {
					return err
				}

				var excluded []string
				var toExclude []string

				if toExcludeArgument := c.String("exclude"); toExcludeArgument != "" {
					toExclude = strings.Split(toExcludeArgument, ",")
				}
				if toExcludeBytes, err := ioutil.ReadAll(os.Stdin); err != nil {
					return err
				} else {
					toExcludeStdin := strings.Split(string(toExcludeBytes), "\n")
					toExclude = append(toExclude, toExcludeStdin...)
				}

				for _, part := range toExclude {
					if n, ok := s.Nodes[extractParenthesis(part)]; ok {
						excluded = append(excluded, n.repr())
					}
				}

				if len(excluded) > 0 {
					for _, nrepr := range excluded {
						id := extractParenthesis(nrepr)
						delete(s.Nodes, id)
					}

					// exclude all relationships for which the two nodes are not present
					for key, r := range s.Rels {
						_, fromexists := s.Nodes[r.From.Id]
						_, toexists := s.Nodes[r.To.Id]
						if !fromexists || !toexists {
							delete(s.Rels, key)
						}
					}
				}

				return tmpl.Execute(os.Stdout, s)
			},
		},
	}

	app.Run(os.Args)
}

const dotTemplate = `
digraph main {
  {{ range .Nodes }}
  n{{ .Id }} [label="{{ .Name }}"]; {{ end }}

  {{ range .Rels }}
  n{{ .From.Id }}->n{{ .To.Id }} [
    label="{{ .Kind }}",
    dir={{ if .Directed }}forward{{ else }}none{{ end }}
  ]; {{ end }}
}
`
