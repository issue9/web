// SPDX-License-Identifier: MIT

package problem

import (
	"sync"

	"github.com/issue9/sliceutil"
)

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

	Params []*InvalidParam `json:"params,omitempty" yaml:"params,omitempty" xml:"params,omitempty"`
}

type InvalidParam struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"i"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Reason  []string `json:"reason" xml:"reason" yaml:"reason"`
}

func NewRFC7807(err FieldErrs) *RFC7807 {
	p := rfc7807ProblemPool.Get().(*RFC7807)
	p.Type = ""
	p.Title = ""
	p.Detail = ""
	p.Instance = ""
	p.Status = 0
	p.Params = p.Params[:0]
	for key, vals := range err {
		p.Params = append(p.Params, &InvalidParam{Name: key, Reason: vals})
	}

	return p
}

func (p *RFC7807) SetType(t string) { p.Type = t }

func (p *RFC7807) SetTitle(title string) { p.Title = title }

func (p *RFC7807) SetDetail(d string) { p.Detail = d }

func (p *RFC7807) SetStatus(s int) { p.Status = s }

func (p *RFC7807) SetInstance(url string) { p.Instance = url }

func (p *RFC7807) AddParam(name string, reason ...string) {
	param, found := sliceutil.At(p.Params, func(pp *InvalidParam) bool { return pp.Name == name })
	if found {
		param.Reason = append(param.Reason, reason...)
	}
	p.Params = append(p.Params, &InvalidParam{Name: name, Reason: reason})
}

func (p *RFC7807) Destroy() { rfc7807ProblemPool.Put(p) }
