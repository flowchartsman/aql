package fmtmsg

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
	ttemplate "text/template"

	"github.com/flowchartsman/aql/parser"
	"github.com/flowchartsman/aql/parser/ast"
)

// Message is a message with position information, such as a ParserMessage or a ParseError
type Message interface {
	Pos() ast.Pos
	Message() string
}

type tmplExecutor interface {
	Execute(io.Writer, any) error
}

// Formatter is used to print query-related messages for display.
type Formatter struct {
	tmpl     tmplExecutor
	maxwidth int
}

// NewFormatter creates a new formatter for printing query-related messages.
func NewFormatter(tmpl string, html bool, maxWidth int) (*Formatter, error) {
	if maxWidth < 0 {
		maxWidth = 0
	}
	var (
		tmplErr error
		t       tmplExecutor
	)
	funcMap := template.FuncMap{
		"pad":     pad,
		"repeat":  strings.Repeat,
		"toupper": strings.ToUpper,
	}
	if html {
		t, tmplErr = template.New("message").Funcs(funcMap).Parse(tmpl)
	} else {
		t, tmplErr = ttemplate.New("message").Funcs(funcMap).Parse(tmpl)
	}
	if tmplErr != nil {
		return nil, fmt.Errorf("template err: %v", tmplErr)
	}
	return &Formatter{
		tmpl:     t,
		maxwidth: maxWidth,
	}, nil
}

// WithQuery returns a [Printer] for messages related to the given query.
func (f *Formatter) WithQuery(query string) *Printer {
	return &Printer{
		formatter: f,
		query:     []rune(query),
	}
}

// Printer prints messages for a given query.
type Printer struct {
	formatter *Formatter
	query     []rune
}

// Fprint formats a query message and writes it to w
func (p *Printer) Fprint(w io.Writer, message Message) {
	if err := p.formatter.tmpl.Execute(w, p.getPrintData(message)); err != nil {
		io.WriteString(w, fmt.Sprintf("**PRINT ERROR: %v", err))
	}
}

// Sprint returns a formatted query message as a string.
func (p *Printer) Sprint(message Message) string {
	var buf bytes.Buffer
	p.Fprint(&buf, message)
	return buf.String()
}

func (p *Printer) getPrintData(m Message) *printData {
	pData := &printData{
		msgType: msgUnknown,
		query:   p.query,
		pos:     m.Pos(),
		lo:      0,
		ro:      len(p.query),
		HasHL:   m.Pos() != ast.NoPosition(),
		Message: m.Message(),
	}
	switch v := m.(type) {
	case *parser.ParseError:
		pData.msgType = msgError
	case *parser.ParserMessage:
		switch v.Type {
		case parser.MsgError:
			pData.msgType = msgError
		case parser.MsgWarning:
			pData.msgType = msgWarning
		case parser.MsgHint:
			pData.msgType = msgInfo
		}
	}
	if p.formatter.maxwidth > 0 {
		truncate(pData, p.formatter.maxwidth)
	}
	return pData
}

type msgType string

const (
	msgUnknown msgType = ""
	msgInfo    msgType = "info"
	msgWarning msgType = "warning"
	msgError   msgType = "error"
)

type printData struct {
	msgType msgType
	query   []rune
	Message string
	HasHL   bool
	pos     ast.Pos
	lo      int
	ro      int
}

func (m *printData) Type() string {
	return string(m.msgType)
}

func (m *printData) HLLen() int {
	return m.pos.Len
}

func (m *printData) HLStrOffset() int {
	offset := m.pos.Offset - m.lo
	if m.lo > 0 {
		offset += 3
	}
	return offset
}

func (m *printData) Query() string {
	if m.lo == 0 && m.ro == len(m.query) {
		return string(m.query)
	}
	var sb strings.Builder
	if m.lo > 0 {
		sb.WriteString("...")
	}
	sb.WriteString(string(m.query[m.lo:m.ro]))
	if m.ro < (len(m.query)) {
		sb.WriteString("...")
	}
	return sb.String()
}

func (m *printData) QueryPre() string {
	if !m.HasHL {
		return m.Query()
	}
	var sb strings.Builder
	if m.lo > 0 {
		sb.WriteString("...")
	}
	sb.WriteString(string(m.query[m.lo:m.pos.Offset]))
	return sb.String()
}

func (m *printData) QueryHL() string {
	if !m.HasHL {
		return ""
	}
	return string(m.query[m.pos.Offset : m.pos.Offset+m.pos.Len])
}

func (m *printData) QueryPost() string {
	if !m.HasHL {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(string(m.query[m.pos.Offset+m.pos.Len : m.ro]))
	if m.ro < (len(m.query)) {
		sb.WriteString("...")
	}
	return sb.String()
}

const terminalTmpl = `
{{- toupper .Type}}: {{.Message}}
{{- if .HasHL }}
    {{.Query}}
    {{pad .HLStrOffset}}{{repeat "^" .HLLen}}
{{end -}}
`

// NewTerminalFormatter is a formatter for printing terminal-formatted messages, along with a marker line showing the location referred to by the message.
func NewTerminalFormatter(maxLen int) *Formatter {
	p, _ := NewFormatter(terminalTmpl, false, maxLen)
	return p
}
