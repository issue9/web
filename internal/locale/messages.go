// SPDX-License-Identifier: MIT

package locale

import (
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

type (
	Messages struct {
		XMLName  struct{}     `xml:"messages" json:"-" yaml:"-"`
		Language language.Tag `xml:"language" json:"language" yaml:"language"`
		Messages []Message    `xml:"message" json:"messages" yaml:"messages"`
	}

	// Message 单条消息
	Message struct {
		Key     string `xml:"key" json:"key" yaml:"key"`
		Message Text   `xml:"message" json:"message" yaml:"message"`
	}

	Text struct {
		Msg    string  `xml:"msg,omitempty" json:"msg,omitempty"  yaml:"msg,omitempty"`
		Select *Select `xml:"select,omitempty" json:"select,omitempty" yaml:"select,omitempty"`
		Vars   []*Var  `xml:"var,omitempty" json:"vars,omitempty" yaml:"vars,omitempty"`
	}

	Select struct {
		Args   int      `xml:"args" json:"args" yaml:"args"`
		Format string   `xml:"format,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
		Cases  []string `xml:"cases" json:"cases" yaml:"cases"` // TODO 必须偶数
	}

	Var struct {
		Name   string  `xml:"name" json:"name" yaml:"name"`
		Select *Select `xml:"select" json:"select" yaml:"select"`
	}
)

func (s *Select) message() catalog.Message {
	cases := make([]interface{}, 0, len(s.Cases))
	for _, c := range s.Cases {
		cases = append(cases, c)
	}
	return plural.Selectf(s.Args, s.Format, cases...)
}

func (v *Var) message() catalog.Message {
	return catalog.Var(v.Name, v.Select.message())
}
