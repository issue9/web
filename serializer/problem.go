// SPDX-License-Identifier: MIT

package serializer

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"sync"

	"github.com/issue9/errwrap"
	"gopkg.in/yaml.v3"
)

var problemPool = &sync.Pool{New: func() any {
	return &Problem{
		Keys: make([]string, 5),
		Vals: make([]any, 5),
	}
}}

const problemPoolMaxSize = 20

const problemXMLNS = "urn:ietf:rfc:7807"

// Problem 表示错误时返回给用户的数据结构
//
// 这是对 [RFC7807] 数据格式的定义，目前仅有 JSON 和 XML 的格式定义，
// 其它格式，可能需要用户在各自的 [MarshalFunc] 方法自行实现。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type Problem struct {
	// NOTE: Problem 不能定义为接口，否则在 MarshalFunc 判断类型时，
	// 可能会将部分无意实现该接口的数据按 Problem 进行序列化。

	// NOTE: map 本身是无序的，每次输出的结果，顺序会不同，不方便测试。
	// 且无法复用，采用双数组的形式，可以通过 sync.Pool 复用对象。

	// NOTE: 需要序列化的内容必须为公开类型，否则 GOB 无法序列化。

	Keys []string
	Vals []any
}

// InvalidParam 表示参数验证错误
//
// 这不是 RFC7807 的一部分，但是比较常用。
type InvalidParam struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"i"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Reason  string   `json:"reason" xml:"reason" yaml:"reason"`
}

type xmlProblemEntry struct {
	XMLName xml.Name
	Value   any `xml:",chardata"`
}

type xmlProblemObjectEntry struct {
	XMLName xml.Name
	Value   any
}

func NewProblem() *Problem {
	p := problemPool.Get().(*Problem)
	p.Keys = p.Keys[:0]
	p.Vals = p.Vals[:0]
	return p
}

func (p *Problem) Destroy() {
	if len(p.Keys) < problemPoolMaxSize {
		problemPool.Put(p)
	}
}

func (p *Problem) Set(key string, val any) {
	p.Keys = append(p.Keys, key)
	p.Vals = append(p.Vals, val)
}

func (p *Problem) MarshalJSON() ([]byte, error) {
	b := errwrap.Buffer{}
	b.WByte('{')
	for index, key := range p.Keys {
		b.WByte('"').WString(key).WString(`":`)

		v, err := json.Marshal(p.Vals[index])
		if err != nil {
			return nil, err
		}
		b.WBytes(v).WByte(',')
	}
	b.Truncate(b.Len() - 1)
	b.WByte('}')

	return b.Bytes(), b.Err
}

func (p *Problem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "problem"
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: problemXMLNS})
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for index, key := range p.Keys {
		val := p.Vals[index]
		v := reflect.ValueOf(val)
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		var err error
		if k := v.Kind(); k <= reflect.Complex128 || k == reflect.String {
			err = e.Encode(xmlProblemEntry{XMLName: xml.Name{Local: key}, Value: val})
		} else if k == reflect.Array || k == reflect.Slice {
			s := xml.StartElement{Name: xml.Name{Local: key}}
			if err = e.EncodeToken(s); err == nil {
				for i := 0; i < v.Len(); i++ {
					if err = e.EncodeElement(v.Index(i).Interface(), xml.StartElement{Name: xml.Name{Local: "i"}}); err != nil {
						return err
					}
				}
				err = e.EncodeToken(s.End())
			}
		} else {
			err = e.EncodeElement(val, xml.StartElement{Name: xml.Name{Local: key}})
		}
		if err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (p *Problem) MarshalYAML() (any, error) {
	n := &yaml.Node{Kind: yaml.MappingNode}

	for index, key := range p.Keys {
		v := &yaml.Node{}
		if err := v.Encode(p.Vals[index]); err != nil {
			return nil, err
		}

		n.Content = append(n.Content, &yaml.Node{
			Value: key,
			Kind:  yaml.ScalarNode,
		}, v)
	}

	return n, nil
}
