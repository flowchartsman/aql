package aqlgraph

import "text/template"

const (
	colorAnd  = `#f59722`
	colorOr   = `#f4e452`
	colorNot  = `#ba4932`
	colorExpr = `#00507c`
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
	Props  []NodeProp
	Values []NodeVal
	Left   string
	Right  string
}

type NodeProp struct {
	Name   string
	Values []NodeVal
}

type NodeVal struct {
	Val   string
	Color string
}

var labelTmpl = template.Must(template.New("").Parse(`<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="2">
<TR>
	<TD COLSPAN="2" CELLPADDING="8">{{ .Field }}</TD>
</TR>
{{- range .Props}}
<TR>
	<TD CELLPADDING="0" BORDER="1">{{ .Name |html }}</TD>
	<TD CELLPADDING="0" BORDER="0">
		<TABLE BORDER="0" CELLPADDING="2" CELLSPACING="0">
{{- range .Values}}
			<TR>
				<TD BORDER="1"><FONT FACE="monospace">{{ . |html }}</FONT></TD>
			</TR>
{{- end }}
		</TABLE>
	</TD>
</TR>
{{end}}
{{- if .Values}}
<TR>
<TD COLSPAN="2">Values</TD>
</TR>
{{- range .Values}}
<TR>
<TD COLSPAN="2" BGCOLOR="{{ . Color }}>{{ .Val | html}}</TD>
</TR>
{{- end}}
{{- end}}
</TABLE>`))
