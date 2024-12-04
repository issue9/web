// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"github.com/issue9/query/v3"

	"github.com/issue9/web"
)

// Tag 关联指定的标签
func (o *Operation) Tag(tag ...string) *Operation {
	o.Tags = append(o.Tags, tag...)
	return o
}

func (o *Operation) Desc(summary, description web.LocaleStringer) *Operation {
	o.Summary = summary
	o.Description = description
	return o
}

func (o *Operation) Server(url string, desc web.LocaleStringer, vars ...*ServerVariable) *Operation {
	s := &Server{
		URL:         url,
		Description: desc,
		Variables:   vars,
	}
	if err := s.valid(); err != nil {
		panic(err)
	}

	o.Servers = append(o.Servers, s)
	return o
}

func (o *Operation) buildParameter(name, typ string, desc web.LocaleStringer, f func(*Parameter)) *Parameter {
	p := &Parameter{
		Name:        name,
		Description: desc,
		Required:    true,
		Schema:      &Schema{Type: typ},
	}
	if f != nil {
		f(p)
	}

	if err := p.valid(true); err != nil {
		panic(err)
	}

	return p
}

// Path 指定路径参数
//
// name 参数名称，如果参数带了类型限定，比如 /path/{id:digit} 等，需要带上类型限定符；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
//
// NOTE: 当同一个路径包含不同的请求方法时，只需要定义其一个请求方法中的路径参数即可，
// 会自动应用到所有的请求方法。
func (o *Operation) Path(name, typ string, desc web.LocaleStringer, f func(*Parameter)) *Operation {
	// TODO 如果支持泛型方法，typ 可以由泛型类型获得

	p := o.buildParameter(name, typ, desc, f)
	p.Required = true
	o.Paths = append(o.Paths, p)
	return o
}

// PathID 指定类型为大于 0 的路径参数
func (o *Operation) PathID(name string, desc web.LocaleStringer) *Operation {
	return o.Path(name, TypeInteger, desc, func(p *Parameter) {
		p.Schema.Minimum = 1
	})
}

func (o *Operation) PathRef(ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.paths[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.Paths = append(o.Paths, &Parameter{
		Ref:      &Ref{Ref: ref, Summary: summary, Description: description},
		Required: true,
	})
	return o
}

// Query 指定一个查询参数
func (o *Operation) Query(name, typ string, desc web.LocaleStringer, f func(*Parameter)) *Operation {
	o.Queries = append(o.Queries, o.buildParameter(name, typ, desc, f))
	return o
}

func (o *Operation) QueryRef(ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.queries[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.Queries = append(o.Queries, &Parameter{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}})
	return o
}

// QueryObject 从参数 o 中获取相应的查询参数
//
// 对于 obj 的要求与 [web.Context.QueryObject] 是相同的。
// 如果参数 obj 非空的，那么该非空字段同时也作为该查询参数的默认值。
// f 是对每个字段的修改，可以为空，其原型为
//
//	func(p *Parameter)
//
// 可通过 p.Name 确定的参数名称
func (o *Operation) QueryObject(obj any, f func(*Parameter)) *Operation {
	return o.queryObject(reflect.ValueOf(obj), f)
}

func (o *Operation) queryObject(v reflect.Value, f func(*Parameter)) *Operation {
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("t 必须得是 struct 类型")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		vt := v.Field(i)

		if ft.Anonymous {
			o.queryObject(vt, f)
			continue
		}

		if !ft.IsExported() {
			continue
		}
		name, _, _ := getTagName(ft, query.Tag)
		if name == "" {
			name = ft.Name
		}

		var desc web.LocaleStringer
		if ft.Tag != "" {
			if c := ft.Tag.Get(CommentTag); c != "" {
				desc = web.Phrase(c)
			}
		}

		p := &Parameter{Name: name, Description: desc} // comment 提取的内容作用在 Parameter 上，而不是关联的 Schema 上
		q := func(p *Parameter) {
			if f != nil {
				f(p)
			}
			if err := p.valid(true); err != nil {
				panic(err)
			}
			o.Queries = append(o.Queries, p)
		}

		p.Schema = &Schema{}
		if !vt.IsZero() {
			p.Schema.Default = vt.Interface()
		}
		schemaFromType(nil, ft.Type, true, "", p.Schema)
		if !p.Schema.isBasicType() {
			panic("不支持复杂类型")
		}
		q(p)
	}

	return o
}

// Header 添加报头
//
// name 报头名称；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
func (o *Operation) Header(name, typ string, desc web.LocaleStringer, f func(*Parameter)) *Operation {
	o.Headers = append(o.Headers, o.buildParameter(name, typ, desc, f))
	return o
}

func (o *Operation) HeaderRef(ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.headers[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.Headers = append(o.Headers, &Parameter{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}})
	return o
}

// Cookie 添加 Cookie
//
// name Cookie 的名称；
// typ 表示该参数在 json schema 中的类型；
// desc 对该参数的表述；
// f 如果 typ 无法描述该参数的全部特征，那么可以使用 f 对该类型进行修正，否则为空；
func (o *Operation) Cookie(name, typ string, desc web.LocaleStringer, f func(*Parameter)) *Operation {
	o.Cookies = append(o.Cookies, o.buildParameter(name, typ, desc, f))
	return o
}

func (o *Operation) CookieRef(ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.cookies[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.Cookies = append(o.Cookies, &Parameter{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}})
	return o
}

// Body 从 body 参数中获取请求内容的类型
//
// f 如果不为空，则要以对根据 body 生成的对象做二次修改。
func (o *Operation) Body(body any, ignorable bool, desc web.LocaleStringer, f func(*Request)) *Operation {
	req := &Request{
		Ignorable:   ignorable,
		Body:        o.d.newSchema(body),
		Description: desc,
	}
	if f != nil {
		f(req)
	}

	if err := req.valid(true); err != nil {
		panic(err)
	}

	o.RequestBody = req
	return o
}

func (o *Operation) BodyRef(ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.requests[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.RequestBody = &Request{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}}
	return o
}

// Response 从 resp 参数中获取返回对象的类型
//
// f 如果不为空，则要以对根据 resp 生成的对象做二次修改。
func (o *Operation) Response(status string, resp any, desc web.LocaleStringer, f func(*Response)) *Operation {
	r := &Response{Description: desc}

	if resp != nil {
		r.Body = o.d.newSchema(resp)
	}

	if f != nil {
		f(r)
	}

	if err := r.valid(true); err != nil {
		panic(err)
	}

	o.Responses[status] = r
	return o
}

func (o *Operation) ResponseRef(status, ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.responses[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	o.Responses[status] = &Response{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}}
	return o
}

// Response200 相当于 o.Response("200", resp, nil, nil)
func (o *Operation) Response200(resp any) *Operation {
	return o.Response("200", resp, nil, nil)
}

// ResponseEmpty 相当于 o.ResponseRef(status, EmptyResponseRef, nil, nil)
func (o *Operation) ResponseEmpty(status string) *Operation {
	return o.ResponseRef(status, EmptyResponseRef, nil, nil)
}

// CallbackRef 引用 components 中定义的回调对象
func (o *Operation) CallbackRef(name, ref string, summary, description web.LocaleStringer) *Operation {
	if _, found := o.d.components.callbacks[ref]; !found {
		panic(fmt.Sprintf("未找到引用 %s", ref))
	}

	if o.Callbacks == nil {
		o.Callbacks = make(map[string]*Callback, 1)
	}
	o.Callbacks[name] = &Callback{Ref: &Ref{
		Ref:         ref,
		Summary:     summary,
		Description: description,
	}}

	return o
}

// Callback 定义回调对象
func (o *Operation) Callback(name, path, method string, f func(*Operation)) *Operation {
	if o.Callbacks == nil {
		o.Callbacks = make(map[string]*Callback, 1)
	}

	c, found := o.Callbacks[name]
	if !found {
		c = &Callback{
			Callback: make(map[string]*PathItem, 1),
		}
		o.Callbacks[name] = c
	}

	item, found := c.Callback[path]
	if !found {
		item = &PathItem{}
		c.Callback[path] = item
	}

	opt, found := item.Operations[method]
	if !found {
		opt = &Operation{d: o.d}
	}

	if f != nil {
		f(opt)
	}

	return o
}

// API 提供用于声明 openapi 文档的中间件
//
// 用户可通过 f 方法提供的参数 o 对接口数据进行更改。
// 对于 [Operation] 的更改，可以直接操作字段，也可以通过其提供的方法进行更改。
// 两者稍有区别，前者不会对数据进行验证。
func (d *Document) API(f func(o *Operation)) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		if pattern != "" && method != "" &&
			(d.enableHead || method != http.MethodHead) &&
			(d.enableOptions || method != http.MethodOptions) {
			o := &Operation{
				d:         d,
				Responses: make(map[string]*Response, len(d.responses)+1), // 必然存在的字段，直接初始化了。
			}
			f(o)

			d.addOperation(method, pattern, router, o)
			d.last = time.Now()
		}
		return next
	})
}

func (d *Document) addOperation(method, pattern, _ string, opt *Operation) {
	pathParams := getPathParams(pattern)
	for _, p := range opt.Paths {
		if slices.Index(pathParams, p.Name) < 0 {
			panic(fmt.Sprintf("路径参数 %s 不存在于路径", p.Name))
		}
	}

	if d.paths == nil {
		d.paths = make(map[string]*PathItem, 50)
	}

	opt.addToComponents(d.components)

	for _, ref := range d.headers { // 添加公共报头的定义
		opt.Headers = append(opt.Headers, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for _, ref := range d.cookies { // 添加公共 Cookie 的定义
		opt.Cookies = append(opt.Cookies, &Parameter{
			Ref: &Ref{Ref: ref},
		})
	}
	for status, ref := range d.responses {
		if _, found := opt.Responses[status]; !found {
			opt.Responses[status] = &Response{Ref: &Ref{Ref: ref}}
		}
	}
	if len(opt.Responses) == 0 {
		panic("至少需要指定一个返回对象")
	}

	for _, t := range opt.Tags {
		if slices.IndexFunc(d.tags, func(elem *tag) bool { return elem.name == t }) < 0 {
			d.tags = append(d.tags, &tag{name: t})
		}
	}

	if item, found := d.paths[pattern]; !found {
		item = &PathItem{
			Operations: make(map[string]*Operation, 3),
			Paths:      opt.Paths,
		}
		opt.Paths = nil

		item.Operations[method] = opt
		d.paths[pattern] = item
	} else {
		if _, found = item.Operations[method]; found {
			panic(fmt.Sprintf("已经存在 %s:%s 的定义", method, pattern))
		}
		if len(item.Paths) == 0 && len(opt.Paths) > 0 {
			item.Paths = opt.Paths
			opt.Paths = nil
		}
		item.Operations[method] = opt
	}
}
