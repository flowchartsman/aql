package aqlgraph

import (
	"strconv"

	"github.com/flowchartsman/aql/parser"
	"github.com/flowchartsman/aql/parser/ast"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Graph represents the graphical representation of an AQL query.
type Graph struct {
	ng *namegraph
}

type namegraph struct {
	*cgraph.Graph
	n namer
}

func (n *namegraph) NextNode() *cgraph.Node {
	newnode, err := n.Graph.CreateNode(n.n.next())
	if err != nil {
		panic("could not create new node: " + err.Error())
	}
	return newnode
}

type namer struct {
	id int
}

func (n *namer) next() string {
	name := strconv.Itoa(n.id)
	n.id++
	return name
}

func QueryGraph(query string) (*Graph, error) {
	p := parser.NewParser()
	root, err := p.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	gv := graphviz.New()
	g, err := gv.Graph(graphviz.Directed)
	g.SetLabelLocation(cgraph.TopLocation).SetLabelJust(cgraph.CenteredJust).SetLabel("query")
	if err != nil {
		return nil, err
	}
	ng := &namegraph{
		Graph: g,
	}
	if err := grecurse(ng, root); err != nil {
		return nil, err
	}
	return &Graph{ng}, nil
}

func grecurse(graph *namegraph, node ast.Node) error {
}

func astToGraph(graph *namegraph, anode ast.Node) *cgraph.Node {
	n := graph.NextNode()
	switch a := anode.(type) {
	case *ast.AndNode:
	case *ast.OrNode:
	case *ast.ExprNode:
	case *ast.NotNode:
	case *ast.SubdocNode:
	}
}
