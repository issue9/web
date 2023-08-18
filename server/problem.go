// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"sync"

	"github.com/issue9/errwrap"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/logs"
)

const (
	problemPoolMaxSize  = 30 // problem.Fields + problem.Params 少于此值才会回收。
	problemParamsKey    = "params"
	rfc8707XMLNamespace = "urn:ietf:rfc:7807"
)

const (
	typeIndex int = iota
	titleIndex
	detailIndex
	statusIndex
	fixedSize
)

var (
	problemPool = &sync.Pool{
		New: func() any {
			return &Problem{
				Fields: []ProblemField{{Key: "type"}, {Key: "title"}, {Key: "detail"}, {Key: "status"}},
			}
		},
	}

	filterProblemPool = &sync.Pool{New: func() any { return &FilterProblem{} }}
)

type (
	// Problem 根据 [RFC7807] 实现向用户反馈非正常状态的信息
	//
	// [MarshalFunc] 的实现者，可能需要对 Problem 进行处理以便输出更加友好的格式。
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
	Problem struct {
		status int

		// 默认的字段
		//
		// 其中前四个固定为 type,title,detail,status。通过 with 添加的字段跟随其后。
		Fields []ProblemField

		// 表示用户提交字段的错误信息
		Params []problemParam
	}

	ProblemField struct {
		Key   string // 字段名
		Value any    // 对应的值
	}

	problemParam struct {
		Name   string // 出错字段的名称
		Reason string // 出错信息
	}

	problemEntry struct {
		XMLName xml.Name
		Value   any `xml:",chardata"`
	}

	// FilterProblem 处理由过滤器生成的各错误
	FilterProblem struct {
		exitAtError bool
		ctx         *Context
		p           *Problem
	}
)

func newProblem() *Problem {
	p := problemPool.Get().(*Problem)

	if p.Fields != nil {
		p.Fields = p.Fields[:fixedSize]
	}

	if p.Params != nil {
		p.Params = p.Params[:0]
	}

	return p
}

func (p *Problem) init(id, title, detail string, status int) *Problem {
	p.Fields[typeIndex].Value = id
	p.Fields[titleIndex].Value = title
	p.Fields[detailIndex].Value = detail
	p.Fields[statusIndex].Value = status
	p.status = status
	return p
}

func (p *Problem) Apply(ctx *Context) {
	ctx.Render(p.status, p, true)
	if len(p.Fields)+len(p.Params) < problemPoolMaxSize {
		problemPool.Put(p)
	}
}

// WithParam 添加具体的错误字段及描述信息
//
// name 为字段名称；reason 为该字段的错误信息；
func (p *Problem) WithParam(name string, reason string) *Problem {
	if _, found := sliceutil.At(p.Params, func(pp problemParam, _ int) bool { return pp.Name == name }); found {
		panic("已经存在")
	}
	p.Params = append(p.Params, problemParam{Name: name, Reason: reason})
	return p
}

// WithField 添加新的输出字段
func (p *Problem) WithField(key string, val any) *Problem {
	if sliceutil.Exists(p.Fields, func(e ProblemField, _ int) bool { return e.Key == key }) || key == problemParamsKey {
		panic("存在同名的参数")
	}
	p.Fields = append(p.Fields, ProblemField{Key: key, Value: val})
	return p
}

// AddProblem 添加新的错误代码
func (srv *Server) AddProblem(id string, status int, title, detail localeutil.LocaleStringer) *Server {
	srv.problems.Add(id, status, title, detail)
	return srv
}

// VisitProblems 遍历错误代码
//
// visit 签名：
//
//	func(prefix, id string, status int, title, detail localeutil.LocaleStringer)
//
// prefix 用户设置的前缀，可能为空值；
// id 为错误代码，不包含前缀部分；
// status 该错误代码反馈给用户的 HTTP 状态码；
// title 错误代码的简要描述；
// detail 错误代码的明细；
func (srv *Server) VisitProblems(visit func(prefix, id string, status int, title, detail localeutil.LocaleStringer)) {
	srv.problems.Visit(visit)
}

// Problem 返回批定 id 的错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 并赋值给返回对象；
func (ctx *Context) Problem(id string) *Problem { return ctx.initProblem(newProblem(), id) }

func (ctx *Context) initProblem(p *Problem, id string) *Problem {
	sp := ctx.Server().problems.Problem(id)
	pp := ctx.LocalePrinter()
	return p.init(sp.Type, sp.Title.LocaleString(pp), sp.Detail.LocaleString(pp), sp.Status)
}

// Error 将 err 输出到 ERROR 通道并尝试以指定 id 的 [Problem] 返回
//
// 如果 id 为空，尝试以下顺序获得值：
//   - err 是否是由 [web.NewHTTPError] 创建，如果是则采用 err.Status 取得 ID 值；
//   - 采用 [problems.ProblemInternalServerError]；
func (ctx *Context) Error(err error, id string) *Problem {
	if id == "" {
		var herr *errs.HTTP
		if errors.As(err, &herr) {
			id = problems.ID(herr.Status)
			err = herr.Message
		} else {
			id = problems.ProblemInternalServerError
		}
	}

	ctx.Logs().NewRecord(logs.Error).DepthError(2, err)
	return ctx.Problem(id)
}

func (ctx *Context) NotFound() *Problem { return ctx.Problem(problems.ProblemNotFound) }

func (ctx *Context) NotImplemented() *Problem { return ctx.Problem(problems.ProblemNotImplemented) }

// NewFilterProblem 声明用于处理过滤器的错误对象
func (ctx *Context) NewFilterProblem(exitAtError bool) *FilterProblem {
	v := filterProblemPool.Get().(*FilterProblem)
	v.exitAtError = exitAtError
	v.ctx = ctx
	v.p = newProblem()
	ctx.OnExit(func(*Context, int) { filterProblemPool.Put(v) })
	return v
}

func (v *FilterProblem) continueNext() bool { return !v.exitAtError || v.len() == 0 }

func (v *FilterProblem) len() int { return len(v.p.Params) }

// Add 添加一条错误信息
func (v *FilterProblem) Add(name string, reason localeutil.LocaleStringer) *FilterProblem {
	if v.continueNext() {
		return v.add(name, reason)
	}
	return v
}

// AddError 添加一条类型为 error 的错误信息
func (v *FilterProblem) AddError(name string, err error) *FilterProblem {
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		return v.Add(name, ls)
	}
	return v.Add(name, localeutil.Phrase(err.Error()))
}

func (v *FilterProblem) add(name string, reason localeutil.LocaleStringer) *FilterProblem {
	v.p.WithParam(name, reason.LocaleString(v.Context().LocalePrinter()))
	return v
}

// AddFilter 添加由过滤器 f 返回的错误信息
func (v *FilterProblem) AddFilter(f filter.FilterFunc) *FilterProblem {
	if !v.continueNext() {
		return v
	}

	if name, msg := f(); msg != nil {
		v.add(name, msg)
	}
	return v
}

// When 只有满足 cond 才执行 f 中的验证
//
// f 中的 v 即为当前对象；
func (v *FilterProblem) When(cond bool, f func(v *FilterProblem)) *FilterProblem {
	if cond {
		f(v)
	}
	return v
}

// Context 返回关联的 [Context] 实例
func (v *FilterProblem) Context() *Context { return v.ctx }

// Problem 转换成 [Problem] 对象
//
// 如果当前对象没有收集到错误，那么将返回 nil。
func (v *FilterProblem) Problem(id string) *Problem {
	if v == nil || v.len() == 0 {
		return nil
	}
	return v.Context().initProblem(v.p, id)
}

// Problem 的 Marshal 实现

func (p *Problem) MarshalJSON() ([]byte, error) {
	b := errwrap.Buffer{}
	b.WByte('{')

	for _, field := range p.Fields {
		b.WByte('"').WString(field.Key).WString(`":`)

		v, err := json.Marshal(field.Value)
		if err != nil {
			return nil, err
		}
		b.WBytes(v).WByte(',')
	}

	if len(p.Params) > 0 {
		b.WByte('"').WString(problemParamsKey + `":[`)
		for _, param := range p.Params {
			b.WString(`{"name":"`).WString(param.Name).WString(`","reason":"`).WString(param.Reason).WString(`"},`)
		}
		b.Truncate(b.Len() - 1)
		b.WString("],")
	}

	b.Truncate(b.Len() - 1)

	b.WByte('}')

	return b.Bytes(), b.Err
}

func (p *Problem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "problem"
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: rfc8707XMLNamespace})
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for _, field := range p.Fields {
		v := reflect.ValueOf(field.Value)
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		var err error
		if k := v.Kind(); k <= reflect.Complex128 || k == reflect.String {
			err = e.Encode(problemEntry{XMLName: xml.Name{Local: field.Key}, Value: field.Value})
		} else if k == reflect.Array || k == reflect.Slice {
			s := xml.StartElement{Name: xml.Name{Local: field.Key}}
			if err = e.EncodeToken(s); err == nil {
				for i := 0; i < v.Len(); i++ {
					if err = e.EncodeElement(v.Index(i).Interface(), xml.StartElement{Name: xml.Name{Local: "i"}}); err != nil {
						return err
					}
				}
				err = e.EncodeToken(s.End())
			}
		} else {
			err = e.EncodeElement(field.Value, xml.StartElement{Name: xml.Name{Local: field.Key}})
		}
		if err != nil {
			return err
		}
	}

	if len(p.Params) > 0 {
		pStart := xml.StartElement{Name: xml.Name{Local: problemParamsKey}}
		if err := e.EncodeToken(pStart); err != nil {
			return err
		}

		for _, param := range p.Params {
			iStart := xml.StartElement{Name: xml.Name{Local: "i"}}
			if err := e.EncodeToken(iStart); err != nil {
				return err
			}

			if err := e.EncodeElement(param.Name, xml.StartElement{Name: xml.Name{Local: "name"}}); err != nil {
				return err
			}
			if err := e.EncodeElement(param.Reason, xml.StartElement{Name: xml.Name{Local: "reason"}}); err != nil {
				return err
			}

			if err := e.EncodeToken(iStart.End()); err != nil {
				return err
			}
		}

		if err := e.EncodeToken(pStart.End()); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (p *Problem) MarshalForm() ([]byte, error) {
	u := url.Values{}

	for _, f := range p.Fields {
		var val string
		switch v := f.Value.(type) {
		case bool:
			val = strconv.FormatBool(v)
		case int:
			val = strconv.FormatInt(int64(v), 10)
		case int8:
			val = strconv.FormatInt(int64(v), 10)
		case int16:
			val = strconv.FormatInt(int64(v), 10)
		case int32:
			val = strconv.FormatInt(int64(v), 10)
		case int64:
			val = strconv.FormatInt(v, 10)
		case uint:
			val = strconv.FormatUint(uint64(v), 10)
		case uint8:
			val = strconv.FormatUint(uint64(v), 10)
		case uint16:
			val = strconv.FormatUint(uint64(v), 10)
		case uint32:
			val = strconv.FormatUint(uint64(v), 10)
		case uint64:
			val = strconv.FormatUint(v, 10)
		case float32:
			val = strconv.FormatFloat(float64(v), 'f', 2, 32)
		case float64:
			val = strconv.FormatFloat(v, 'f', 2, 32)
		case string:
			val = v
		default: // 其它类型忽略
			continue
		}
		u.Add(f.Key, val)
	}

	for _, param := range p.Params {
		u.Add(problemParamsKey+"."+param.Name, param.Reason)
	}

	return []byte(u.Encode()), nil
}

func (p *Problem) MarshalHTML() (string, any) {
	data := make(map[string]any, len(p.Fields))
	for _, f := range p.Fields {
		data[f.Key] = f.Value
	}

	if len(p.Params) > 0 {
		ps := make(map[string]string, len(p.Params))
		for _, param := range p.Params {
			ps[param.Name] = param.Reason
		}
		data[problemParamsKey] = ps
	}

	return "problem", data
}
