package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/flowchartsman/aql/jsonmatcher"
	"github.com/flowchartsman/aql/parser"
	"github.com/flowchartsman/aql/parser/fmtmsg"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("Usage: aql 'EXPR' <json file>")
	}

	query := os.Args[1]

	if strings.HasSuffix(query, ".aql") {
		qb, err := os.ReadFile(query)
		if err != nil {
			log.Fatal(err)
		}
		query = string(qb)
	}

	printer := fmtmsg.NewTerminalFormatter(-1).WithQuery(query)
	m, err := jsonmatcher.NewMatcher(query)
	if err != nil {
		var parseErr *parser.ParseError
		if errors.As(err, &parseErr) {
			log.Fatal(printer.Sprint(parseErr))
		}
		log.Fatal(err)
	}

	input, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}

	for _, m := range m.Messages() {
		log.Println(printer.Sprint(m))
	}

	result, err := m.Match(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	if stats := m.Stats(); stats != nil {
		statsB, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			log.Fatalf("error marshalling stats: %s", err)
		}
		fmt.Println(string(statsB))
	}
}
