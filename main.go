package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func main() {
	app := cli.NewApp()

	app.Name = "rels"
	app.Description = "manage entities and relationships between them in flat files."
	app.Version = "0.0.1"

	s := state{}

	app.Before = func(c *cli.Context) error {
		here, _ := os.Getwd()
		filepath.Walk(here, func(path string, f os.FileInfo, err error) error {
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
			err = yaml.Unmarshal(contents, &raw)
			if err != nil {
				log.Print("couldn't parse yaml from file: ", path, ". ", err)
				return nil
			}

			var id string
			if fid, ok := raw["id"].(string); ok {
				id = fid
			} else {
				id = strings.Split(filepath.Base(path), ".")[0]
			}

			s.nodes = append(s.nodes, node{
				raw:  raw,
				path: path,
				name: raw["name"].(string),
				id:   id,
			})

			return nil
		})

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "nodes",
			Usage: "list all nodes",
			Action: func(c *cli.Context) error {
				for i, node := range s.nodes {
					fmt.Printf("%d\t%s\t%s\n", i, node.id, node.name)
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
				fmt.Println("completed task: ", c.Args().First())
				return nil
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
	nodes  []node
	schema schema
}

type node struct {
	id       string
	path     string
	raw      interface{}
	name     string
	outgoing []*rel
	incoming []*rel
	neutral  []*rel
}

type rel struct {
	kind    string
	neutral bool
	from    *node
	to      *node
}

type schema struct {
}
