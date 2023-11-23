// SPDX-License-Identifier: MIT

package enum

import (
	"strings"

	"github.com/issue9/errwrap"
)

// Type 枚举类型的数据
type Type struct {
	name     string  // 类型名称
	values   []value // 类型的所有可能值
	receiver string

	type2StringMapName string
	string2TypeMapName string
}

type value struct {
	Name   string // 值名称
	String string // 值对应的字符串值
}

func NewType(t string, vals ...string) *Type {
	has := true
	for _, v := range vals {
		if has = has && strings.HasPrefix(v, t); !has {
			break
		}
	}

	values := make([]value, 0, len(vals))
	if has {
		for _, v := range vals {
			values = append(values, value{Name: v, String: strings.ToLower(strings.TrimPrefix(v, t))})
		}
	} else {
		for _, v := range vals {
			values = append(values, value{Name: v, String: strings.ToLower(v)})
		}
	}

	return &Type{
		name:     t,
		values:   values,
		receiver: string(t[0]),

		type2StringMapName: "_" + t + "ToString",
		string2TypeMapName: "_" + t + "FromString",
	}
}

func (t *Type) dump(buf *errwrap.Buffer) {
	buf.Printf("\n\n//---------------------- %s ----------------------\n\n", t.name)

	// type2StringMap
	buf.Printf("var %s=map[%s]string{\n", t.type2StringMapName, t.name)
	for _, v := range t.values {
		buf.Printf("%s:\"%s\",\n", v.Name, v.String)
	}
	buf.WString("}\n\n")

	// string2TypeMap
	buf.Printf("var %s=map[string]%s{\n", t.string2TypeMapName, t.name)
	for _, v := range t.values {
		buf.Printf("\"%s\":%s,\n", v.String, v.Name)
	}
	buf.WString("}\n\n")

	// String
	buf.WString("// String fmt.Stringer\n").
		Printf("func (%s %s)String()string{\n", t.receiver, t.name).
		Printf("if v,found := %s[%s];found{\n", t.type2StringMapName, t.receiver).
		WString("return v\n").
		WString("}\n").
		Printf(`return fmt.Sprintf("%s(%%d)", %s)`, t.name, t.receiver).WRune('\n').
		WString("}\n\n")

	// TextMarshaler
	buf.WString("// MarshalText encoding.TextMarshaler\n").
		Printf("func(%s %s) MarshalText()([]byte,error){\n", t.receiver, t.name).
		Printf("if v,found := %s[%s];found{\n", t.type2StringMapName, t.receiver).
		WString("return []byte(v),nil\n").
		WString("}\n").
		Printf(`return []byte(fmt.Sprintf("%s(%%d)", %s)),locales.ErrInvalidValue()`, t.name, t.receiver).WRune('\n').
		WString("}\n\n")

	// Parse
	buf.Printf("// Parse%s 将字符串 v 解析为 %s 类型\n", t.name, t.name).
		Printf("func Parse%s(v string)(%s,error){\n", t.name, t.name).
		Printf("if t,found := %s[v];found{\n", t.string2TypeMapName).
		WString("return t,nil\n").
		WString("}\n").
		WString(`return 0,locales.ErrInvalidValue()`).WRune('\n').
		WString("}\n\n")

	// TextUnmarshaler
	buf.WString("// UnmarshalText encoding.TextUnmarshaler\n").
		Printf("func(%s *%s) UnmarshalText(p []byte)(error){\n", t.receiver, t.name).
		Printf("tmp,err :=Parse%s(string(p))\n", t.name).
		WString("if err==nil{\n").
		Printf("*%s=tmp\n", t.receiver).
		WString("}\n").
		WString("return err\n").
		WString("}\n\n")

	// IsValid
	buf.WString("// IsValid 验证该状态值是否有效\n").
		Printf("func(%s %s)IsValid()bool{\n", t.receiver, t.name).
		Printf("_,found :=%s[%s];\n", t.type2StringMapName, t.receiver).
		WString("return found\n").
		WString("}\n\n")

	// Validator
	buf.Printf("func %sValidator(v %s) bool {\n", t.name, t.name).
		WString("return v.IsValid()\n").
		WString("}\n\n")

	// rule
	buf.Printf(`var %sRule = web.NewRule(%sValidator, locales.InvalidValue)`, t.name, t.name)
	buf.WString("\n\n")

	// sliceRule
	buf.Printf(`var %sSliceRule = web.NewSliceRules[%s,[]%s](%sRule)`, t.name, t.name, t.name, t.name)
	buf.WString("\n\n")

	// filter
	buf.Printf(`var %sFilter = web.NewFilter(%sRule)`, t.name, t.name)
	buf.WString("\n\n")

	// sliceFilter
	buf.Printf(`var %sSliceFilter = web.NewFilter(%sSliceRule)`, t.name, t.name)
}
