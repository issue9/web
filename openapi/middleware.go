// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"

	"github.com/issue9/query/v3"

	"github.com/issue9/web"
)

// Tag 关联指定的标签
func (o *Operation) Tag(tag ...string) *Operation {
	o.Tags = append(o.Tags, tag...)
	return o
}

// Path 指定路径参数
//
// name 参数名称，如果参数带了类型限定，比如 /path/{id:digit} 等，需要带上类型限定符；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
func (o *Operation) Path(name, typ string, desc web.LocaleStringer, f func(*Schema)) *Operation {
	// TODO 如果支持泛型方法，typ 可以由泛型类型获得

	s := &Schema{Type: typ}
	if f != nil {
		f(s)
	}

	o.Paths = append(o.Paths, &Parameter{
		Name:        name,
		Description: desc,
		Required:    true,
		Schema:      s,
	})
	return o
}

func (o *Operation) PathRef(ref string) *Operation {
	o.Paths = append(o.Paths, &Parameter{Ref: &Ref{Ref: ref}, Required: true})
	return o
}

// Query 指定一个查询参数
func (o *Operation) Query(name, typ string, desc web.LocaleStringer, f func(*Schema)) *Operation {
	s := &Schema{Type: typ}
	if f != nil {
		f(s)
	}

	return o.query(name, s, desc)
}

func (o *Operation) query(name string, s *Schema, desc web.LocaleStringer) *Operation {
	o.Queries = append(o.Queries, &Parameter{Name: name, Description: desc, Schema: s})
	return o
}

func (o *Operation) QueryRef(ref string) *Operation {
	o.Queries = append(o.Queries, &Parameter{Ref: &Ref{Ref: ref}})
	return o
}

// QueryObject 从参数 o 中获取相应的查询参数
//
// 对于 o 的要求与 [web.Context.QueryObject] 是相同的。
// f 是对每个字段的修改，可以为空，其原型为
//
//	func(p *Parameter)
//
// 可通过 p.Name 确定的参数名称
func (m *Operation) QueryObject(o any, f func(*Parameter)) *Operation {
	t := reflect.TypeOf(o)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("o 必须得是 struct 类型")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, _ := getTagName(field, query.Tag)
		if name == "" {
			name = field.Name
		}

		p := &Parameter{Name: name}
		q := func(p *Parameter) {
			if f != nil {
				f(p)
			}
			m.Queries = append(m.Queries, p)
		}
		switch field.Type.Kind() {
		case reflect.String:
			p.Schema = &Schema{Type: TypeString}
			q(p)
		case reflect.Bool:
			p.Schema = &Schema{Type: TypeBoolean}
			q(p)
		case reflect.Float32, reflect.Float64:
			p.Schema = &Schema{Type: TypeNumber}
			q(p)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p.Schema = &Schema{Type: TypeInteger}
			q(p)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			p.Schema = &Schema{Type: TypeInteger, Minimum: 0}
			q(p)
		case reflect.Array, reflect.Slice:
			p.Schema = &Schema{Type: TypeArray, Items: schemaFromType(m.d, reflect.TypeOf(field.Type.Elem()), false, "")}
			q(p)
		default:
			panic(fmt.Sprintf("查询参数不支持复杂的类型 %v", field.Type.Kind()))
		}
	}

	return m
}

// Header 添加报头
//
// name 报头名称；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
func (o *Operation) Header(name, typ string, desc web.LocaleStringer, f func(*Schema)) *Operation {
	s := &Schema{Type: typ}
	if f != nil {
		f(s)
	}

	o.Headers = append(o.Headers, &Parameter{Name: name, Description: desc, Schema: s})
	return o
}

func (o *Operation) HeaderRef(ref string) *Operation {
	o.Headers = append(o.Headers, &Parameter{Ref: &Ref{Ref: ref}})
	return o
}

// Cookie 添加 Cookie
//
// name Cookie 的名称；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
func (o *Operation) Cookie(name, typ string, desc web.LocaleStringer, f func(*Schema)) *Operation {
	s := &Schema{Type: typ}
	if f != nil {
		f(s)
	}

	o.Cookies = append(o.Cookies, &Parameter{Name: name, Description: desc, Schema: s})
	return o
}

func (o *Operation) CookieRef(ref string) *Operation {
	o.Cookies = append(o.Cookies, &Parameter{Ref: &Ref{Ref: ref}})
	return o
}

// Body 从 body 参数中获取请求内容的类型
//
// f 如果不为空，则要以对根据 body 生成的对象做二次修改。
func (o *Operation) Body(body any, f func(*Request)) *Operation {
	req := &Request{
		Body: o.d.newSchema(reflect.TypeOf(body)),
	}
	if f != nil {
		f(req)
	}

	o.RequestBody = req
	return o
}

func (o *Operation) BodyRef(ref string) *Operation {
	o.RequestBody = &Request{Ref: &Ref{Ref: ref}}
	return o
}

// Response 从 resp 参数中获取返回对象的类型
//
// f 如果不为空，则要以对根据 resp 生成的对象做二次修改。
func (o *Operation) Response(status int, resp any, desc web.LocaleStringer, f func(*Response)) *Operation {
	r := &Response{
		Description: desc,
		Body:        o.d.newSchema(reflect.TypeOf(resp)),
	}
	if f != nil {
		f(r)
	}

	o.Responses[status] = r
	return o
}

func (o *Operation) ResponseRef(status int, ref string) *Operation {
	o.Responses[status] = &Response{Ref: &Ref{Ref: ref}}
	return o
}

// API 提供用于声明 openapi 文档的中间件
func (d *Document) API(f func(o *Operation)) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		if !d.disable {
			o := &Operation{
				d:         d,
				Responses: make(map[int]*Response, 1), // 必然存在的字段，直接初始化了。
			}
			f(o)

			d.addOperation(method, pattern, router, o)
		}
		return next
	})
}

func (d *Document) addOperation(method, pattern, _ string, opt *Operation) {
	if (!d.enableHead && method == http.MethodHead) || (!d.enableOptions && method == http.MethodOptions) || pattern == "" {
		return
	}

	pathParams := getPathParams(pattern)
	for _, p := range opt.Paths {
		if slices.Index(pathParams, p.Name) < 0 {
			panic(fmt.Sprintf("路径参数 %s 不存在于路径", p.Name))
		}
	}

	if d.paths == nil {
		d.paths = make(map[string]*PathItem, 50)
	}

	opt.addComponents(d.components)

	for _, ref := range d.headers {
		opt.Headers = append(opt.Headers, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for _, ref := range d.cookies {
		opt.Cookies = append(opt.Cookies, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for status, ref := range d.responses {
		if _, found := opt.Responses[status]; !found {
			opt.Responses[status] = &Response{Ref: &Ref{Ref: ref}}
		}
	}

	for _, t := range opt.Tags {
		if slices.IndexFunc(d.tags, func(elem *tag) bool { return elem.name == t }) < 0 {
			d.tags = append(d.tags, &tag{name: t})
		}
	}

	if item, found := d.paths[pattern]; !found {
		item = &PathItem{Operations: make(map[string]*Operation, 3)}
		item.Operations[method] = opt
		d.paths[pattern] = item
	} else {
		if _, found = item.Operations[method]; found {
			panic(fmt.Sprintf("已经存在 %s:%s 的定义", method, pattern))
		}
		item.Operations[method] = opt
	}
}
