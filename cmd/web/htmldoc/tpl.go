// SPDX-License-Identifier: MIT

package htmldoc

const defaultStyle = `
:root {
	--color: black;
	--bg: white;
}
@media (prefers-color-scheme: dark) {
	:root {
		--color: white;
		--bg: black;
	}
}
table {
	width: 100%;
	border-collapse: collapse;
	border: 1px solid var(--color);
	text-align: left;
}
th {
	text-align: left;
}
tr {
	border-bottom: 1px solid var(--color);
}
td {
	padding-left: 5px;
	padding-right: 3px;
}

body {
	color: var(--color);
	background: var(--bg);
}`

const tpl = `<!DOCTYPE html>
<html lang="{{.Lang}}">
	<head>
		<title>{{.Title}}</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		{{- if .Style}}
		<style>
		{{.Style}}
		</style>
		{{- end}}
	</head>
	<body>
		{{- if .Header}}
			{{.Header}}
		{{else}}
			<h1>{{.Title}}</h1>
			{{- if .Desc}}<article>{{.Desc}}</article>{{end -}}
		{{end}}

		{{- range .Objects}}
			{{- if .Title}}<h2 id="{{.Title}}">{{.Title}}</h2>{{end -}}
			{{- if .Desc}}<article>{{.Desc}}</article>{{end -}}
			{{if .Items}}
				<table>
					<thead><tr><th>JSON</th><th>YAML</th><th>XML</th><th>{{$.TypeLocale}}</th><th>{{$.DescLocale}}</th><tr></thead>
					<tbody>
					{{range .Items -}}
						<tr><td>{{.JSON}}</td><td>{{.YAML}}</td><td>{{.XML}}</td><td>{{.Type}}</td><td>{{.Desc}}</td></tr>
					{{- end}}
					</tbody>
				</table>
			{{end}}
		{{- end}}
		{{- if .Footer}}{{.Footer}}{{end}}
	</body>
</html>`
