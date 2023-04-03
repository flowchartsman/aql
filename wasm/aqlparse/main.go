//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/flowchartsman/aql/parser"
)

func parseAQL(query string) (string, []*parser.ParserMessage, error) {
	visitor := parser.NewMessageVisitor(warningVisitor)
	root, perr := parser.ParseQuery(query, parser.Visitors(visitor))
	if perr != nil {
		return "", nil, perr
	}
	tree, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("marshal error: %v", err)
	}
	return string(tree), visitor.Messages(), nil
}

func errConvert(err error, input string) map[string]any {
	if pe, ok := err.(*parser.ParseError); ok {
		msg := pe.Msg
		if strings.HasPrefix(msg, "no match found") {
			msg = "incomplete query"
			pe.Position.Offset = len(input) - 1
			pe.Position.Len = 1
		}
		return map[string]any{
			"type":  "error",
			"msg":   msg,
			"start": pe.Position.Offset,
			"end":   pe.Position.Offset + pe.Position.Len,
		}
	} else {
		return map[string]any{
			"msg": err.Error(),
		}
	}
}

func msgConvert(msg *parser.ParserMessage, input string) map[string]any {
	var msgType string
	switch msg.Type {
	case parser.MsgHint:
		msgType = "hint"
	case parser.MsgWarning:
		msgType = "warning"
	case parser.MsgError:
		msgType = "error"
	}
	return map[string]any{
		"type":  msgType,
		"msg":   msg.Msg,
		"start": msg.Position.Offset,
		"end":   msg.Position.Offset + msg.Position.Len,
	}
}

func pWrap(this js.Value, args []js.Value) any {
	input := args[0].String()
	result := map[string]any{
		"ast": "",
	}
	messages := []any{}
	root, pmessages, err := parseAQL(input)
	if err != nil {
		messages = append(messages, errConvert(err, input))
	}
	if len(pmessages) > 0 {
		for _, m := range pmessages {
			messages = append(messages, msgConvert(m, input))
		}
	}
	if root != "" {
		result["ast"] = root
	}
	result["messages"] = messages
	return result
}

func main() {
	// js.Global().Set("parseAQL", parseWrapper())
	js.Global().Set("parseAQL", js.FuncOf(pWrap))
	js.Global().Get("notifyBrowser").Invoke()
	select {}
}
