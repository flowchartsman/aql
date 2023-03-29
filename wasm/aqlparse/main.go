//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/flowchartsman/aql/parser"
)

func parseWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return jsResp("", errors.New("Not enough arguments"))
		}
		input := args[0].String()
		return jsResp(parseAQL(input))
	})
}

func jsResp(encoded string, err error) map[string]any {
	var errstr string
	if err != nil {
		errstr = err.Error()
	}
	return map[string]any{
		"error":   errstr,
		"encoded": encoded,
	}
}

func parseAQL(query string) (string, error) {
	root, perr := parser.ParseQuery(query)
	if perr != nil {
		return "", perr
	}
	tree, err := json.Marshal(root)
	if err != nil {
		return "", fmt.Errorf("marshal error: %v", err)
	}
	return string(tree), nil
}

func errConvert(err error) map[string]any {
	if pe, ok := err.(*parser.ParseError); ok {
		return map[string]any{
			"msg":   pe.Message(),
			"start": pe.Position.Offset,
			"end":   pe.Position.Offset + pe.Position.Len,
		}
	} else {
		return map[string]any{
			"msg": err.Error(),
		}
	}
}

func pWrap(this js.Value, args []js.Value) any {
	input := args[0].String()
	result := map[string]any{
		"errors": []any{},
		"ast":    "",
	}
	root, err := parseAQL(input)
	if err != nil {
		errors := []any{}
		errors = append(errors, errConvert(err))
		result["errors"] = errors
	} else {
		b, _ := json.MarshalIndent(root, "", "  ")
		result["ast"] = string(b)
	}
	return result
}

func main() {
	// js.Global().Set("parseAQL", parseWrapper())
	js.Global().Set("parseAQL", js.FuncOf(pWrap))
	js.Global().Get("notifyBrowser").Invoke()
	select {}
}
