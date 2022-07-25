// SPDX-License-Identifier: MIT

package serializer

import (
	"sync"

	"github.com/issue9/validation"
)

var rfc7807ProblemPool = &sync.Pool{New: func() any {
	return &RFC7807Problem{}
}}

type FieldErrs = validation.LocaleMessages

// Problem 错误信息对象需要实现的接口
//
// 字段参考了 [RFC7807]，但是没有固定具体的呈现方式，
// 用户可以自定义具体的渲染方法。可以使用 [RFC7807Problem]。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type Problem interface {
	GetType() string
	SetType(string)

	GetTitle() string
	SetTitle(string)

	GetDetail() string
	SetDetail(string)

	GetStatus() int
	SetStatus(int)

	GetInstance() string
	SetInstance(string)

	// Destroy 销毁当前对象
	//
	// 如果当前对象采用了类似 sync.Pool 的技术对内容进行了保留，
	// 那么可以在此方法中调用 [sync.Pool.Put] 方法返回给对象池。
	// 否则的话可以实现为一个空方法即可。
	Destroy()
}

// RFC7807Problem [Problem] 接口的 [RFC7807] 标准实现
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type RFC7807Problem struct {
	XMLName  struct{} `json:"-" yaml:"-" xml:"urn:ietf:rfc:7807 problem"`
	Type     string   `json:"type" yaml:"type" xml:"type"`
	Title    string   `json:"title" yaml:"title" xml:"title"`
	Detail   string   `json:"detail,omitempty" yaml:"detail,omitempty" xml:"detail,omitempty"`
	Status   int      `json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	Instance string   `json:"instance,omitempty" yaml:"instance,omitempty" xml:"instance,omitempty"`
}

// InvalidParamsProblem 这是表示参数错误的 RFC7807 扩展
type InvalidParamsProblem struct {
	RFC7807Problem
	InvalidParams []*InvalidParam `json:"invalid-params" yaml:"invalid-params" xml:"invalid-params"`
}

type InvalidParam struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"i"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Reason  string   `json:"reason" xml:"reason" yaml:"reason"`
}

func NewInvalidParamsProblem(err FieldErrs) Problem {
	p := &InvalidParamsProblem{}
	for key, vals := range err {
		for _, val := range vals {
			p.InvalidParams = append(p.InvalidParams, &InvalidParam{Name: key, Reason: val})
		}
	}
	return p
}

func NewRFC7807Problem() Problem {
	p := rfc7807ProblemPool.Get().(*RFC7807Problem)
	p.Type = ""
	p.Title = ""
	p.Detail = ""
	p.Instance = ""
	p.Status = 0
	return p
}

func (p *RFC7807Problem) GetType() string { return p.Type }

func (p *RFC7807Problem) SetType(t string) { p.Type = t }

func (p *RFC7807Problem) GetTitle() string { return p.Title }

func (p *RFC7807Problem) SetTitle(title string) { p.Title = title }

func (p *RFC7807Problem) GetDetail() string { return p.Detail }

func (p *RFC7807Problem) SetDetail(d string) { p.Detail = d }

func (p *RFC7807Problem) GetStatus() int { return p.Status }

func (p *RFC7807Problem) SetStatus(s int) { p.Status = s }

func (p *RFC7807Problem) GetInstance() string { return p.Instance }

func (p *RFC7807Problem) SetInstance(url string) { p.Instance = url }

func (p *RFC7807Problem) Destroy() { rfc7807ProblemPool.Put(p) }
