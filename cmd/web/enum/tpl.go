// SPDX-License-Identifier: MIT

package enum

const tpl = `// {{.FileHeader}}

package {{.Package}}

import(
	"fmt"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
)

{{range .Types}}
//---------------------{{.Name}}------------------------

var {{.Type2StringMap}} = map[{{.Name}}] string {
	{{- range .Values}}
	{{.Name}}:"{{.String}}",
	{{- end}}
}

var {{.String2TypeMap}} = map[string]{{.Name}} {
	{{- range .Values}}
	"{{.String}}":{{.Name}},
	{{- end}}
}

// String fmt.Stringer
func({{.Receiver}} {{.Name}})String()string {
	if v, found := {{.Type2StringMap}}[{{.Receiver}}]; found {
		return v
	}
	return fmt.Sprintf("{{.Name}}(%d)", {{.Receiver}})
}

func Parse{{.Name}}(v string)({{.Name}},error){
	if t,found := {{.String2TypeMap}}[v];found{
		return t, nil
	}
	return 0, locales.ErrInvalidValue()
}

// MarshalText encoding.TextMarshaler
func({{.Receiver}} {{.Name}}) MarshalText()([]byte,error){
	if v, found := {{.Type2StringMap}}[{{.Receiver}}]; found {
		return []byte(v),nil
	}
	return nil, locales.ErrInvalidValue()
}

// UnmarshalText encoding.TextUnmarshaler
func({{.Receiver}} *{{.Name}}) UnmarshalText(p []byte)(error){
	tmp,err :=Parse{{.Name}}(string(p))
	if err==nil{
		*{{.Receiver}}=tmp
	}
	return err
}

func({{.Receiver}} {{.Name}})IsValid()bool{
	_,found :={{.Type2StringMap}}[{{.Receiver}}];
	return found
}

func {{.Name}}Validator(v {{.Name}}) bool {return v.IsValid()}

var(
	{{.Name}}Rule = web.NewRule({{.Name}}Validator, locales.InvalidValue)

	{{.Name}}SliceRule = web.NewSliceRules[{{.Name}},[]{{.Name}}]({{.Name}}Rule)

	{{.Name}}Filter = web.NewFilter({{.Name}}Rule)

	{{.Name}}SliceFilter = web.NewFilter({{.Name}}SliceRule)
)

{{end}}
`
