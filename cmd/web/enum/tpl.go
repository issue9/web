// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package enum

const tpl = `// {{.FileHeader}}

package {{.Package}}

import(
	"fmt"
	"database/sql/driver"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

{{range .Enums}}

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

func({{.Receiver}} *{{.Name}})Scan(src any)error {
	if src == nil {
		return locales.ErrInvalidValue()
	}

	var val string
	switch v := src.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	case []rune:
		val = string(v)
	default:
		return locales.ErrInvalidValue()
	}

	v,err :=Parse{{.Name}}(val)
	if err!=nil{
		return err
	}

	*{{.Receiver}} = v
	return nil
}

func({{.Receiver}} {{.Name}})Value()(driver.Value,error) {
	v,err := {{.Receiver}}.MarshalText()
	if err!=nil{return nil,err}
	return string(v),nil
}

func {{.Name}}Validator(v {{.Name}}) bool {return v.IsValid()}

var(
	{{.Name}}Rule = filter.V({{.Name}}Validator, locales.InvalidValue)

	{{.Name}}SliceRule = filter.SV[[]{{.Name}}]({{.Name}}Validator, locales.InvalidValue)

	{{.Name}}Filter = filter.NewBuilder({{.Name}}Rule)

	{{.Name}}SliceFilter = filter.NewBuilder({{.Name}}SliceRule)
)

//---------------------end {{.Name}}--------------------

{{end}}
`
