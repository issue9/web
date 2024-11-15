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

type APIMiddleware struct {
	d *Document
	o *Operation
}

// Path 指定路径参数
func (m *APIMiddleware) Path(name, typ string, desc web.LocaleStringer) *APIMiddleware {
	// TODO 如果支持泛型方法，typ 可以由泛型类型获得

	if m.o.Paths == nil {
		m.o.Paths = []*Parameter{}
	}
	m.o.Paths = append(m.o.Paths, &Parameter{
		Name:        name,
		Description: desc,
		Required:    true,
		Schema:      &Schema{Type: typ},
	})
	return m
}

func (m *APIMiddleware) PathRef(ref string) *APIMiddleware {
	if m.o.Paths == nil {
		m.o.Paths = []*Parameter{}
	}
	m.o.Paths = append(m.o.Paths, &Parameter{Ref: &Ref{Ref: ref}, Required: true})
	return m
}

// Query 指定一个查询参数
func (m *APIMiddleware) Query(name, typ string, desc web.LocaleStringer) *APIMiddleware {
	return m.query(name, &Schema{Type: typ}, desc)
}

func (m *APIMiddleware) query(name string, s *Schema, desc web.LocaleStringer) *APIMiddleware {
	if m.o.Queries == nil {
		m.o.Queries = []*Parameter{}
	}
	m.o.Queries = append(m.o.Queries, &Parameter{Name: name, Description: desc, Schema: s})
	return m
}

func (m *APIMiddleware) QueryRef(ref string) *APIMiddleware {
	if m.o.Queries == nil {
		m.o.Queries = []*Parameter{}
	}
	m.o.Queries = append(m.o.Queries, &Parameter{Ref: &Ref{Ref: ref}})
	return m
}

// QueryObject 从参数 o 中获取相应的查询参数
//
// 对于 o 的要求与 [web.Context.QueryObject] 是相同的。
func (m *APIMiddleware) QueryObject(o any) *APIMiddleware {
	t := reflect.TypeOf(o)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("o 必须得是 struct 类型")
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _ := getTagName(f, query.Tag)
		if name == "" {
			name = f.Name
		}

		switch f.Type.Kind() {
		case reflect.String:
			m.Query(name, TypeString, nil)
		case reflect.Bool:
			m.Query(name, TypeBoolean, nil)
		case reflect.Float32, reflect.Float64:
			m.Query(name, TypeNumber, nil)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m.Query(name, TypeInteger, nil)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			m.query(name, &Schema{Type: TypeInteger, Minimum: 0}, nil)
		case reflect.Array, reflect.Slice:
			m.query(name, &Schema{Type: TypeArray, Items: schemaFromType(m.d, reflect.TypeOf(f.Type.Elem()), false, "")}, nil)
		default:
			panic(fmt.Sprintf("查询参数不支持复杂的类型 %v", f.Type.Kind()))
		}
	}

	return m
}

func (m *APIMiddleware) Header(name, typ string, desc web.LocaleStringer) *APIMiddleware {
	if m.o.Headers == nil {
		m.o.Headers = []*Parameter{}
	}
	m.o.Headers = append(m.o.Headers, &Parameter{Name: name, Description: desc, Schema: &Schema{Type: typ}})
	return m
}

func (m *APIMiddleware) HeaderRef(ref string) *APIMiddleware {
	if m.o.Headers == nil {
		m.o.Headers = []*Parameter{}
	}
	m.o.Headers = append(m.o.Headers, &Parameter{Ref: &Ref{Ref: ref}})
	return m
}

func (m *APIMiddleware) Cookie(name, typ string, desc web.LocaleStringer) *APIMiddleware {
	if m.o.Cookies == nil {
		m.o.Cookies = []*Parameter{}
	}
	m.o.Cookies = append(m.o.Cookies, &Parameter{Name: name, Description: desc, Schema: &Schema{Type: typ}})
	return m
}

func (m *APIMiddleware) CookieRef(ref string) *APIMiddleware {
	if m.o.Cookies == nil {
		m.o.Cookies = []*Parameter{}
	}
	m.o.Cookies = append(m.o.Cookies, &Parameter{Ref: &Ref{Ref: ref}})
	return m
}

// Body 从 body 参数中获取请求内容的类型
func (m *APIMiddleware) Body(body any) *APIMiddleware {
	m.o.RequestBody = &Request{
		Body: m.d.newSchema(reflect.TypeOf(body)),
	}
	return m
}

func (m *APIMiddleware) BodyRef(ref string) *APIMiddleware {
	m.o.RequestBody = &Request{Ref: &Ref{Ref: ref}}
	return m
}

// Response 从 resp 参数中获取返回对象的类型
func (m *APIMiddleware) Response(status int, resp any, desc web.LocaleStringer) *APIMiddleware {
	m.o.Responses[status] = &Response{
		Description: desc,
		Body:        m.d.newSchema(reflect.TypeOf(resp)),
	}
	return m
}

func (m *APIMiddleware) ResponseRef(status int, ref string) *APIMiddleware {
	m.o.Responses[status] = &Response{Ref: &Ref{Ref: ref}}
	return m
}

func (m *APIMiddleware) Middleware(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
	if !m.d.disable {
		m.d.addOperation(method, pattern, router, m.o)
	}
	return next
}

// API 提供用于声明 openapi 文档的中间件
//
// NOTE: 与 [Document.Operation] 功能是相同的，但是 API 作了大部分简化，
// 且可以通过对象实例直接获取各个字段名称，不需要像 [Operation] 那样需要手动填定对象的字段。
func (d *Document) API(tag ...string) *APIMiddleware {
	return &APIMiddleware{o: &Operation{
		Tags:      tag,
		Responses: make(map[int]*Response, 1), // 必然存在的字段，直接初始化了。
	}, d: d}
}

// Operation 提供根据的 [Operation] 生成 openapi 文档的中间件
func (d *Document) Operation(o *Operation) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		if !d.disable {
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
