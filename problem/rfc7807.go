// SPDX-License-Identifier: MIT

package problem

import "sync"

var rfc7807ProblemPool = &sync.Pool{New: func() any {
	return &RFC7807{}
}}

// RFC7807 [Problem] 接口的 [RFC7807] 标准实现
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type RFC7807 struct {
	XMLName  struct{} `json:"-" yaml:"-" xml:"urn:ietf:rfc:7807 problem"`
	Type     string   `json:"type" yaml:"type" xml:"type"`
	Title    string   `json:"title" yaml:"title" xml:"title"`
	Detail   string   `json:"detail,omitempty" yaml:"detail,omitempty" xml:"detail,omitempty"`
	Status   int      `json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	Instance string   `json:"instance,omitempty" yaml:"instance,omitempty" xml:"instance,omitempty"`
}

// InvalidParamsProblem 这是表示参数错误的 RFC7807 扩展
type InvalidParamsProblem struct {
	RFC7807
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
	p := rfc7807ProblemPool.Get().(*RFC7807)
	p.Type = ""
	p.Title = ""
	p.Detail = ""
	p.Instance = ""
	p.Status = 0
	return p
}

func (p *RFC7807) GetType() string { return p.Type }

func (p *RFC7807) SetType(t string) { p.Type = t }

func (p *RFC7807) GetTitle() string { return p.Title }

func (p *RFC7807) SetTitle(title string) { p.Title = title }

func (p *RFC7807) GetDetail() string { return p.Detail }

func (p *RFC7807) SetDetail(d string) { p.Detail = d }

func (p *RFC7807) GetStatus() int { return p.Status }

func (p *RFC7807) SetStatus(s int) { p.Status = s }

func (p *RFC7807) GetInstance() string { return p.Instance }

func (p *RFC7807) SetInstance(url string) { p.Instance = url }

func (p *RFC7807) Destroy() { rfc7807ProblemPool.Put(p) }
