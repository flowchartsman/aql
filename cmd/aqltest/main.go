package main

import (
	"fmt"
	"log"
	"os"

	"github.com/flowchartsman/aql/jsonquery"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("Usage: aql 'EXPR' <json file>")
	}

	q, err := jsonquery.NewQuerier(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}

	result, err := q.Match(file)
	if err != nil {
		log.Fatalf("error running query: %s", parser.GetPrintableError(os.Args[1], err))
	}
	fmt.Println(result)
}
