// SPDX-License-Identifier: MIT

package problem

import (
	"sync"

	"github.com/issue9/sliceutil"
)

var rfc7807ProblemPool = &sync.Pool{New: func() any {
	return &rfc7807{}
}}

// RFC7807Builder [BuildFunc] 的 [RFC7807] 标准实现
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func RFC7807Builder(id, title, detail string, status int) Problem {
	if id == "" {
		id = "about:blank"
	}

	p := rfc7807ProblemPool.Get().(*rfc7807)
	p.Type = id
	p.Title = title
	p.Detail = detail
	p.Instance = ""
	p.OriginStatus = status
	p.Params = p.Params[:0]

	return p
}

type rfc7807 struct {
	XMLName      struct{} `json:"-" yaml:"-" xml:"urn:ietf:rfc:7807 problem"`
	Type         string   `json:"type" yaml:"type" xml:"type"`
	Title        string   `json:"title" yaml:"title" xml:"title"`
	Detail       string   `json:"detail,omitempty" yaml:"detail,omitempty" xml:"detail,omitempty"`
	OriginStatus int      `json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	Instance     string   `json:"instance,omitempty" yaml:"instance,omitempty" xml:"instance,omitempty"`

	Params []*invalidParam `json:"params,omitempty" yaml:"params,omitempty" xml:"params,omitempty"`
}

type invalidParam struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"i"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Reason  []string `json:"reason" xml:"reason" yaml:"reason"`
}

func (p *rfc7807) Status() int { return p.OriginStatus }

func (p *rfc7807) SetInstance(url string) { p.Instance = url }

func (p *rfc7807) AddParam(name string, reason ...string) {
	param, found := sliceutil.At(p.Params, func(pp *invalidParam) bool { return pp.Name == name })
	if found {
		param.Reason = append(param.Reason, reason...)
	}
	p.Params = append(p.Params, &invalidParam{Name: name, Reason: reason})
}

func (p *rfc7807) Destroy() { rfc7807ProblemPool.Put(p) }
