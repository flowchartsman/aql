//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/flowchartsman/aql/parser"
)

var p = parser.NewParser()

func parseAQL(query string) (string, error) {
	root, err := p.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("invalid query: %v", err)
	}
	tree, err := json.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("marshal error: %v", err)
	}
	return string(tree), nil
}

func main() {
	js.Global().Set("parseAQL", parseWrapper())
	select {}
}

func parseWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return jsResp("", errors.New("Not enough arguments"))
		}
		input := args[0].String()
		return jsResp(parseAQL(input))
	})
}

func jsResp(encoded string, err error) map[string]interface{} {
	var errstr string
	if err != nil {
		errstr = err.Error()
	}
	return map[string]interface{}{
		"error":   errstr,
		"encoded": encoded,
	}
}
