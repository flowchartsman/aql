package aqlgraph

import "text/template"

const (
	colorAnd = `#f59722`
	colorOr  = `#f4e452`
	colorNot = `#ba4932`
	// colorExpr = `#00507c`
	colorExpr = `#427baa`
	colorSub  = `#000000`

	colorFloat  = `#24b581`
	colorInt    = `#007a55`
	colorString = `#cccccc`
	colorNet    = `#d8c732`
	colorBool   = `#ffffff`
	colorRegex  = `#ca074c`
	colorTime   = `#f48544`
)

type htmlNode struct {
	Field  string
	Props  []NodeProp
	Values []NodeVal
}

type NodeProp struct {
	Name  string
	Value string
}

type NodeVal struct {
	ValStr  string
	ValType string
	Color   string
}

var labelTmpl = template.Must(template.New("").Parse(`<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="2">
<TR>
	<TD COLSPAN="2" CELLPADDING="8">{{ .Field }}</TD>
</TR>
{{- range .Props}}
<TR>
	<TD CELLPADDING="0" BORDER="1">{{ .Name |html }}</TD>
	<TD CELLPADDING="0" BORDER="1">{{ .Value|html }}</TD>
</TR>
{{- end }}
{{- if .Values }}
<TR>
	<TD COLSPAN="2">Values</TD>
</TR>
{{- range .Values}}
<TR>
	<TD BORDER="1" BGCOLOR="{{ .Color }}"><I>{{ .ValType }}</I></TD>
	<TD BORDER="1" BGCOLOR="#FFFFFF">{{ .ValStr }}</TD>
</TR>
{{- end }}
{{- end }}
</TABLE>
`))
