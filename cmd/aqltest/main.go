package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/flowchartsman/aql/parser"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("Usage: aql 'EXPR'")
	}
	log.Printf("received string: [%s]\n", os.Args[1])
	// got, err := ParseReader("", strings.NewReader(os.Args[1]), Debug(false))
	got, err := parser.ParseQuery(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// dumpNode(got.(*Node), 0)
	spew.Dump(got)
	fmt.Printf("%#v\n", got)
}
