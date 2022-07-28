// SPDX-License-Identifier: MIT

package problem

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strconv"
	"sync"

	"github.com/issue9/errwrap"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/server/response"
)

const rfc8707XMLNamespace = "urn:ietf:rfc:7807"

const rfc8707PoolMaxSize = 10

var rfc7807ProblemPool = &sync.Pool{New: func() any {
	return &rfc7807{}
}}

// RFC7807Builder [BuildFunc] 的 [RFC7807] 标准实现
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func RFC7807Builder(id, title string, status int) Problem {
	if id == "" {
		id = "about:blank"
	}

	p := rfc7807ProblemPool.Get().(*rfc7807)
	p.typ = id
	p.title = title
	p.status = status
	p.pKeys = p.pKeys[:0]
	p.pReasons = p.pReasons[:0]
	p.keys = p.keys[:0]
	p.vals = p.vals[:0]

	return p
}

type rfc7807 struct {
	typ    string
	title  string
	status int

	// params
	pKeys    []string
	pReasons []string

	// with
	keys []string
	vals []any
}

type rfc8707Entry struct {
	XMLName xml.Name
	Value   any `xml:",chardata"`
}

func (p *rfc7807) AddParam(name string, reason string) Problem {
	if _, found := sliceutil.At(p.pKeys, func(pp string) bool { return pp == name }); found {
		panic("已经存在")
	}
	p.pKeys = append(p.pKeys, name)
	p.pReasons = append(p.pReasons, reason)

	return p
}

func (p *rfc7807) With(key string, val any) Problem {
	p.keys = append(p.keys, key)
	p.vals = append(p.vals, val)
	return p
}

func (p *rfc7807) MarshalJSON() ([]byte, error) {
	b := errwrap.Buffer{}
	b.WByte('{')

	b.WString(`"type":"`).WString(p.typ).WString(`",`).
		WString(`"title":"`).WString(p.title).WString(`",`).
		WString(`"status":`).WString(strconv.Itoa(p.status)).WByte(',')

	if len(p.pKeys) > 0 {
		b.WString(`"params":[`)
		for index, key := range p.pKeys {
			b.WString(`{"name":"`).WString(key).WString(`","reason":"`).WString(p.pReasons[index]).WString(`"},`)
		}
		b.Truncate(b.Len() - 1)
		b.WString("],")
	}

	for index, key := range p.keys {
		b.WByte('"').WString(key).WString(`":`)

		v, err := json.Marshal(p.vals[index])
		if err != nil {
			return nil, err
		}
		b.WBytes(v).WByte(',')
	}

	b.Truncate(b.Len() - 1)
	b.WByte('}')

	return b.Bytes(), b.Err
}

func (p *rfc7807) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "problem"
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: rfc8707XMLNamespace})
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if err := e.EncodeElement(p.typ, xml.StartElement{Name: xml.Name{Local: "type"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(p.title, xml.StartElement{Name: xml.Name{Local: "title"}}); err != nil {
		return err
	}
	if err := e.EncodeElement(p.status, xml.StartElement{Name: xml.Name{Local: "status"}}); err != nil {
		return err
	}

	if len(p.pKeys) > 0 {
		pStart := xml.StartElement{Name: xml.Name{Local: "params"}}
		if err := e.EncodeToken(pStart); err != nil {
			return err
		}

		for index, key := range p.pKeys {
			iStart := xml.StartElement{Name: xml.Name{Local: "i"}}
			if err := e.EncodeToken(iStart); err != nil {
				return err
			}

			if err := e.EncodeElement(key, xml.StartElement{Name: xml.Name{Local: "name"}}); err != nil {
				return err
			}
			if err := e.EncodeElement(p.pReasons[index], xml.StartElement{Name: xml.Name{Local: "reason"}}); err != nil {
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

	for index, key := range p.keys {
		val := p.vals[index]
		v := reflect.ValueOf(val)
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		var err error
		if k := v.Kind(); k <= reflect.Complex128 || k == reflect.String {
			err = e.Encode(rfc8707Entry{XMLName: xml.Name{Local: key}, Value: val})
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

func (p *rfc7807) Apply(ctx response.Context) {
	if err := ctx.Marshal(p.status, p); err != nil {
		ctx.Logs().ERROR().Error(err)
	}

	if len(p.keys) < rfc8707PoolMaxSize && len(p.pKeys) < rfc8707PoolMaxSize {
		rfc7807ProblemPool.Put(p)
	}
}
