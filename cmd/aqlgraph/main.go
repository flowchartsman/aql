package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
		log.Fatal(err)
	}
	makeGraph(os.Stdout, query, rootNode)
}

const premable = `digraph G {
	labelloc=top;
	node [style="filled" fontname = "Helvetica"];
	label=%q;
	root[shape="point"]
	root->0
`

func makeGraph(w io.StringWriter, label string, rootNode *parser.Node) {
	w.WriteString(fmt.Sprintf(premable, label))
	id := 0
	graphNode(w, rootNode, &id)
	w.WriteString(`}`)
}

const (
	colorAnd  = `#D3696C`
	colorOr   = `#D4996A`
	colorTerm = `#679B99`
)

func graphNode(w io.StringWriter, node *parser.Node, id *int) {
	var (
		label string
		color string
	)
	nodeShape := "box"
	switch node.NodeType {
	case parser.NodeAnd:
		label = `"AND"`
		color = colorAnd
	case parser.NodeOr:
		label = `"OR"`
		color = colorOr
	case parser.NodeTerminal:
		color = colorTerm
		nodeShape = "plain"
		var sb strings.Builder
		sb.WriteString(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="2"><TR><TD COLSPAN="2" CELLPADDING="8">TERM</TD></TR><TR><TD CELLPADDING="0" BORDER="0"><TABLE BORDER="0" CELLPADDING="2" CELLSPACING="0">`)
		sb.WriteString(fmt.Sprintf(`<TR><TD BORDER="1">field</TD><TD BORDER="1">%s</TD></TR>`, node.Comparison.Field))
		sb.WriteString(fmt.Sprintf(`<TR><TD BORDER="1">op</TD><TD BORDER="1">%s</TD></TR>`, node.Comparison.Op))
		if len(node.Comparison.Values) == 1 {
			sb.WriteString(fmt.Sprintf(`<TR><TD BORDER="1">value</TD><TD BORDER="1"><FONT FACE="monospace">%s</FONT></TD></TR>`, node.Comparison.Values[0]))
		} else {
			sb.WriteString(`<TR><TD BORDER="1">values</TD><TD CELLPADDING="0" BORDER="0"><TABLE BORDER="0" CELLPADDING="2" CELLSPACING="0">`)
			for _, val := range node.Comparison.Values {
				sb.WriteString(fmt.Sprintf(`<TR><TD BORDER="1"><FONT FACE="monospace">%s</FONT></TD></TR>`, val))
			}
			sb.WriteString(`</TABLE></TD></TR>`)
		}
		sb.WriteString(fmt.Sprintf(`<TR><TD BORDER="1">negated</TD><TD BORDER="1">%v</TD></TR>`, node.Comparison.Negated))
		sb.WriteString(`</TABLE></TD></TR></TABLE>>`)
		label = sb.String()
	}
	w.WriteString("\t")
	w.WriteString(fmt.Sprintf(`%d[shape="%s" fillcolor="%s" label=%s]`, *id, nodeShape, color, label))
	w.WriteString("\n")
	myID := *id
	if node.Left != nil {
		w.WriteString(fmt.Sprintf("\t%d->%d [penwidth=3]\n", myID, *id+1))
		*id++
		graphNode(w, node.Left, id)
	}
	if node.Right != nil {
		w.WriteString(fmt.Sprintf("\t%d->%d\n", myID, *id+1))
		*id++
		graphNode(w, node.Right, id)
	}
}
