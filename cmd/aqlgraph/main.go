package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowchartsman/aql/aqlgraph"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <aql query string> [output file]", os.Args[0])
	}
	query := os.Args[1]
	outputFile := "query_graph.svg"
	ext := ".svg"
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
		ext = strings.ToLower(filepath.Ext(outputFile))
	}
	graph, err := aqlgraph.QueryGraph(query)
	if err != nil {
		log.Fatal(err)
	}

	switch ext {
	case ".svg":
		err = graph.OutputSVG(outputFile)
	case ".dot":
		err = graph.OutputDOT(outputFile)
	default:
		log.Fatalf("unknown file format: %s", ext)
	}

	if err != nil {
		log.Fatal(err)
	}
}
