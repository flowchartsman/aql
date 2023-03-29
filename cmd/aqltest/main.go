package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/flowchartsman/aql/jsonmatcher"
	"github.com/flowchartsman/aql/parser"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("Usage: aql 'EXPR' <json file>")
	}

	m, messages, perr := jsonmatcher.NewMatcher(os.Args[1])
	if perr != nil {
		log.Fatalf("error running query: %s", parser.PrettyErr(os.Args[1], perr))
	}

	if len(messages) > 0 {
		for _, m := range parser.PrettyMessages(os.Args[1], messages...) {
			log.Println(m)
		}
	}

	input, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}

	result, err := m.Match(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	stats := m.Stats()
	b, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling stats: %s", err)
	}
	fmt.Println(string(b))
}
