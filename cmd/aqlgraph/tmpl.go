package main

import "text/template"

var graphTmpl = template.Must(template.New("").Parse(
	`digraph G {
	labelloc=top;
	center=true;
	nodesep=0.5;
	margin=1;
	node [style="filled" fontname = "Helvetica"];
	label="{{ .Title | js }}";
	root[shape="point"]
	{{- range .Nodes }}
		{{- if .Props }}
	{{ .ID }}[shape={{ .Shape }} fillcolor="{{ .Color }}" label=<
<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="2">
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
					<TD BORDER="1"><FONT FACE="monospace">{{ . }}</FONT></TD>
				</TR>
{{- end }}
			</TABLE>
		</TD>
	</TR>
{{- end}}
</TABLE>>]
{{- else }}
	{{ .ID }}[shape={{ .Shape }} {{if (eq .Shape "circle")}}fixedsize=true height=0.4 fontsize=10{{end}} fillcolor="{{ .Color }}" label="{{ .Label }}"]
{{- end }}
{{- end }}
	root->0
	{{- range .Nodes }}
	{{- if .Left }}
	{{ .ID }}->{{ .Left }} [penwidth=3]
	{{- end}}
	{{- if .Right}}
	{{ .ID }}->{{ .Right }}
	{{- end}}
	{{- end }}
}`))
