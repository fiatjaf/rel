# rel

a command line utility to create and manage personal graphs, then write them to dot and make images with graphviz.

![](screencast.gif)

## install

Builds for all systems are available at https://gobuilder.me/github.com/fiatjaf/rel

or you can `go get github.com/fiatjaf/rel`

## how it works

`rel` is basically a file editor. On every run it will read all YAML files in the current directory and use data found on them to build an internal graph representation. When you tell it to add a node it will just create a new YAML file, when you tell it to add a relationship (relationships are called "links", by the way) it will just specify these relationships in the YAML files of the related nodes.

If you want to modify a file's name or add custom metadata, you can just edit the node file (just don't modify the ids). This means you can also save your graph to `git` or do anything else you can do with files.

## usage

* **rel add <node name>** adds a node
* **rel print** shows a prompt with all node names available to tab/autocomplete, then outputs the selected node file contents.
* **rel edit** does the same as `rel print`, but opens an editor for you to directly edit the node file.
* **rel link [--neutral] <relationship kind>** opens a "from", then a "to" prompts from which you can search existing nodes or add new ones, then creates a relationship <from>-><to> with the specified `<kind>`. `--neutral` means the relationship is not directed.
* **rel unlink** opens a prompt with all relationships available, so you can delete some.
* **rel [--json] nodes** lists all nodes sorted by name, `--json` makes it output a JSON array, useful for piping to [jq](https://stedolan.github.io/jq/manual/) and doing advanced filtering.
* **rel [--json] links** lists all relationships.
* **rel dot** outputs a [dot](http://www.graphviz.org/content/dot-language) representation of the graph.
* **rel template --template <file>** renders the given [Go template](https://golang.org/pkg/text/template/) with the data from the graph.

## use cases

 * family trees
 * ?

## this repository traffic statistics

[![](https://ght.trackingco.de/fiatjaf/rel)](https://ght.trackingco.de/)
