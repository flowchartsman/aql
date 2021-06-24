package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/flowchartsman/aql/parser"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("Usage: aqlgraph 'EXPR'")
	}
	query := os.Args[1]
	rootNode, err := parser.ParseQuery(query)
	if err != nil {
		log.Fatal(parser.GetPrintableError(query, err))
	}
	makeGraph(query, rootNode)
}

func makeGraph(title string, rootNode *parser.Node) {
	id := 0
	g := &graph{
		Title: title,
	}
	graphNodes(g, rootNode, &id)
	graphTmpl.Execute(os.Stdout, g)
}

const (
	colorAnd  = `#F0B84E`
	colorOr   = `#DFD3AE`
	colorTerm = `#0E86B0`
	colorNot  = `#D96737`
)

type GraphnodeProp struct {
	Name   string
	Values []string
}

type graphNode struct {
	ID    string
	Shape string
	Field string
	Label string
	Color string
	Props []GraphnodeProp
	Left  string
	Right string
}

func (g *graphNode) addProp(name string, values ...string) {
	g.Props = append(g.Props, GraphnodeProp{Name: name, Values: values})
}

type graph struct {
	Title string
	Nodes []*graphNode
}

func (g *graph) addNode(node *graphNode) {
	g.Nodes = append(g.Nodes, node)
}

func graphNodes(g *graph, node *parser.Node, id *int) {
	current := &graphNode{
		ID: strconv.Itoa(*id),
	}
	switch node.NodeType {
	case parser.NodeAnd:
		current.Label = "AND"
		current.Color = colorAnd
		current.Shape = "box"
	case parser.NodeOr:
		current.Label = "OR"
		current.Color = colorOr
		current.Shape = "box"
	case parser.NodeNot:
		current.Label = "NOT"
		current.Color = colorNot
		current.Shape = "box"
	case parser.NodeTerminal:
		current.Shape = "plain"
		current.Color = colorTerm
		current.Field = strings.Join(node.Comparison.Field, `.`)
		current.addProp("op", node.Comparison.Op)
		current.addProp("values", node.Comparison.Values...)
	}
	g.addNode(current)
	if node.Left != nil {
		*id++
		current.Left = strconv.Itoa(*id)
		if node.Left.Comparison.Negated {
			g.addNode(&graphNode{
				Label: "NOT",
				Color: colorNot,
				Shape: "circle",
				ID:    strconv.Itoa(*id),
				Left:  strconv.Itoa(*id + 1),
			})
			*id++
		}
		graphNodes(g, node.Left, id)
	}
	if node.Right != nil {
		*id++
		current.Right = strconv.Itoa(*id)
		if node.Right.Comparison.Negated {
			g.addNode(&graphNode{
				Label: "NOT",
				Color: colorNot,
				Shape: "circle",
				ID:    strconv.Itoa(*id),
				Right: strconv.Itoa(*id + 1),
			})
			*id++
		}
		graphNodes(g, node.Right, id)
	}
}
