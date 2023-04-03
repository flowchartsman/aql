package aqlgraph

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/flowchartsman/aql/parser"
	"github.com/flowchartsman/aql/parser/ast"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Graph represents the graphical representation of an AQL query.
type Graph struct {
	gv *graphviz.Graphviz
	g  *cgraph.Graph
}

func (g *Graph) OutputSVG(filename string) error {
	return g.gv.RenderFilename(g.g, graphviz.SVG, filename)
}

func (g *Graph) OutputDOT(filename string) error {
	return g.gv.RenderFilename(g.g, "dot", filename)
}

type namegraph struct {
	*cgraph.Graph
	n namer
}

func (n *namegraph) NewNode() *cgraph.Node {
	newnode, err := n.Graph.CreateNode(n.n.next())
	if err != nil {
		panic("could not create new node: " + err.Error())
	}
	return newnode
}

func (n *namegraph) NewEdge(start *cgraph.Node, end *cgraph.Node) *cgraph.Edge {
	newedge, err := n.Graph.CreateEdge(n.n.next(), start, end)
	if err != nil {
		panic("could not create new edge: " + err.Error())
	}
	return newedge
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
	root, err := parser.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	gv := graphviz.New()
	g, err := gv.Graph(graphviz.Name("query"), graphviz.Directed)
	g.SafeSet("fontname", "Courier", "")
	g.SafeSet("fontsize", "10px", "")
	if err != nil {
		return nil, err
	}
	ng := &namegraph{
		Graph: g,
	}
	ng.SetLabel(query).
		SetLabelLocation(cgraph.TopLocation).
		SetLabelJust(cgraph.CenteredJust).
		SetNodeSeparator(0.5).
		SetStyle(cgraph.FilledGraphStyle)
	if err != nil {
		return nil, err
	}
	startNode, err := ng.CreateNode("start")
	if err != nil {
		return nil, err
	}
	startNode.SetShape(cgraph.PointShape)
	rootNode := visitNodes(ng, root)
	firstEdge := ng.NewEdge(startNode, rootNode)
	firstEdge.SetPenWidth(3)
	return &Graph{
		gv: gv,
		g:  ng.Graph,
	}, nil
}

func visitNodes(graph *namegraph, anode ast.Node) *cgraph.Node {
	n := graph.NewNode()
	n.SetStyle(cgraph.FilledNodeStyle)
	n.SafeSet("fontname", "Courier", "")
	n.SafeSet("fontsize", "10px", "")
	var (
		left  ast.Node
		right ast.Node
	)
	switch a := anode.(type) {
	case *ast.NotNode:
		n.SetLabel("NOT")
		n.SetShape(cgraph.BoxShape)
		n.SetFillColor(colorNot)
		left = a.Expr
	case *ast.AndNode:
		n.SetLabel("AND")
		n.SetShape(cgraph.BoxShape)
		n.SetFillColor(colorAnd)
		left = a.Left
		right = a.Right
	case *ast.OrNode:
		n.SetLabel("OR")
		n.SetShape(cgraph.BoxShape)
		n.SetFillColor(colorOr)
		left = a.Left
		right = a.Right
	case *ast.ExprNode:
		n.SetShape(cgraph.PlainShape)
		n.SetFillColor(colorExpr)
		// Populate the labeltable :D
		// Probably want the "EXPR here"
		hn := htmlNode{
			Field: strings.Join(a.Field, "."),
			Props: []NodeProp{
				{
					Name:  "op",
					Value: string(a.Op),
				},
			},
		}
		for _, rv := range a.RVals {
			valColor := ""
			valType := ""
			switch rv.(type) {
			case *ast.FloatVal:
				valColor = colorFloat
				valType = "float"
			case *ast.IntVal:
				valColor = colorInt
				valType = "integer"
			case *ast.StringVal:
				valColor = colorString
				valType = "string"
			case *ast.NetVal:
				valColor = colorNet
				valType = "network"
			case *ast.BoolVal:
				valColor = colorBool
				valType = "boolean"
			case *ast.RegexpVal:
				valColor = colorRegex
				valType = "regexp"
			case *ast.TimeVal:
				valColor = colorTime
				valType = "datetime"
			}
			hn.Values = append(hn.Values, NodeVal{
				ValType: valType,
				ValStr:  rv.String(),
				Color:   valColor,
			})
		}

		// Set the label to html
		var lb bytes.Buffer
		if err := labelTmpl.Execute(&lb, hn); err != nil {
			panic("template error: " + err.Error())
		}

		n.SetLabel(graph.StrdupHTML(lb.String()))
		// case *ast.SubdocNode:
	}
	if left != nil {
		lNode := visitNodes(graph, left)
		ledge := graph.NewEdge(n, lNode)
		ledge.SetPenWidth(3)
	}
	if right != nil {
		lNode := visitNodes(graph, right)
		graph.NewEdge(n, lNode)
	}
	return n
}
