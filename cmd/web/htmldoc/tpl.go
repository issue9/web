// SPDX-License-Identifier: MIT

package htmldoc

const tpl = `<!DOCTYPE html>
<html lang="{{.Lang}}">
	<head>
		<title>{{.Title}}</title>
		<meta charset="utf-8" />
		<style>
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
		}
		</style>
	</head>
	<body>
		<h1>{{.Title}}</h1>
		{{- if .Desc}}<article>{{.Desc}}</article>{{end -}}

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
	</body>
</html>`