// SPDX-License-Identifier: MIT

package serializer

import (
	"sync"

	"github.com/issue9/validation"
)

var standardsProblemPool = &sync.Pool{New: func() any {
	return &StandardsProblem{}
}}

type FieldErrs = validation.LocaleMessages

// Problem [RFC7807] 定义的数据接口
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

// StandardsProblem [RFC7807] 中定义的标准字段
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type StandardsProblem struct {
	XMLName  struct{} `json:"-" yaml:"-" xml:"urn:ietf:rfc:7807 problem"`
	Type     string   `json:"type" yaml:"type" xml:"type"`
	Title    string   `json:"title" yaml:"title" xml:"title"`
	Detail   string   `json:"detail,omitempty" yaml:"detail,omitempty" xml:"detail,omitempty"`
	Status   int      `json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	Instance string   `json:"instance,omitempty" yaml:"instance,omitempty" xml:"instance,omitempty"`
}

// InvalidParamsProblem 这是对 RFC7807 的扩展表示参数验证错误时的对象
type InvalidParamsProblem struct {
	StandardsProblem
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

func NewStandardsProblem() Problem {
	p := standardsProblemPool.Get().(*StandardsProblem)
	p.Type = ""
	p.Title = ""
	p.Detail = ""
	p.Instance = ""
	p.Status = 0
	return p
}

func (p *StandardsProblem) GetType() string { return p.Type }

func (p *StandardsProblem) SetType(t string) { p.Type = t }

func (p *StandardsProblem) GetTitle() string { return p.Title }

func (p *StandardsProblem) SetTitle(title string) { p.Title = title }

func (p *StandardsProblem) GetDetail() string { return p.Detail }

func (p *StandardsProblem) SetDetail(d string) { p.Detail = d }

func (p *StandardsProblem) GetStatus() int { return p.Status }

func (p *StandardsProblem) SetStatus(s int) { p.Status = s }

func (p *StandardsProblem) GetInstance() string { return p.Instance }

func (p *StandardsProblem) SetInstance(url string) { p.Instance = url }

func (p *StandardsProblem) Destroy() { standardsProblemPool.Put(p) }
