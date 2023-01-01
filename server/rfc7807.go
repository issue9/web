// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"net/url"
	"reflect"
	"strconv"
	"sync"

	"github.com/issue9/errwrap"
	"github.com/issue9/sliceutil"
)

const paramsKey = "params"

const rfc8707XMLNamespace = "urn:ietf:rfc:7807"

const rfc8707PoolMaxSize = 10

var rfc7807ProblemPool = &sync.Pool{New: func() any {
	return &rfc7807{
		keys: []string{"type", "title", "status"},
		vals: make([]any, 3),
	}
}}

// RFC7807Builder [BuildProblemFunc] 的 [RFC7807] 标准实现
//
// NOTE: 由于 www-form-urlencoded 对复杂对象的表现能力有限，
// 在此模式下将忽略由 [Problem.With] 添加的复杂类型，只保留基本类型。
//
// 如果是用于 HTML 输出，返回对象实现了 [serializer/html.Marshaler] 接口。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
func RFC7807Builder(id, title string, status int) Problem {
	if id == "" {
		id = aboutBlank
	}

	p := rfc7807ProblemPool.Get().(*rfc7807)
	p.status = status

	p.keys = p.keys[:3]
	p.vals = p.vals[:3]
	// keys 前三个元素为固定值
	p.vals[0] = id
	p.vals[1] = title
	p.vals[2] = status

	p.pKeys = p.pKeys[:0]
	p.pReasons = p.pReasons[:0]

	return p
}

type rfc7807 struct {
	status int

	// with
	keys []string
	vals []any

	// params
	pKeys    []string
	pReasons []string
}

type rfc7807Entry struct {
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
	if sliceutil.Exists(p.keys, func(e string) bool { return e == key }) || key == paramsKey {
		panic("存在同名的参数")
	}
	p.keys = append(p.keys, key)
	p.vals = append(p.vals, val)
	return p
}

func (p *rfc7807) MarshalJSON() ([]byte, error) {
	b := errwrap.Buffer{}
	b.WByte('{')

	for index, key := range p.keys {
		b.WByte('"').WString(key).WString(`":`)

		v, err := json.Marshal(p.vals[index])
		if err != nil {
			return nil, err
		}
		b.WBytes(v).WByte(',')
	}

	if len(p.pKeys) > 0 {
		b.WByte('"').WString(paramsKey + `":[`)
		for index, key := range p.pKeys {
			b.WString(`{"name":"`).WString(key).WString(`","reason":"`).WString(p.pReasons[index]).WString(`"},`)
		}
		b.Truncate(b.Len() - 1)
		b.WString("],")
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

	for index, key := range p.keys {
		val := p.vals[index]
		v := reflect.ValueOf(val)
		for v.Kind() == reflect.Pointer {
			v = v.Elem()
		}

		var err error
		if k := v.Kind(); k <= reflect.Complex128 || k == reflect.String {
			err = e.Encode(rfc7807Entry{XMLName: xml.Name{Local: key}, Value: val})
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

	if len(p.pKeys) > 0 {
		pStart := xml.StartElement{Name: xml.Name{Local: paramsKey}}
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

	return e.EncodeToken(start.End())
}

func (p *rfc7807) MarshalForm() ([]byte, error) {
	u := url.Values{}

	for index, key := range p.keys {
		var val string
		switch v := p.vals[index].(type) {
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
		u.Add(key, val)
	}

	for index, key := range p.pKeys {
		prefix := paramsKey + "[" + strconv.Itoa(index) + "]."
		u.Add(prefix+"name", key)
		u.Add(prefix+"reason", p.pReasons[index])
	}

	return []byte(u.Encode()), nil
}

func (p *rfc7807) MarshalHTML() (string, any) {
	data := make(map[string]any, len(p.keys)+1)
	for index, key := range p.keys {
		data[key] = p.vals[index]
	}

	if len(p.pKeys) > 0 {
		ps := make(map[string]string, len(p.pKeys))
		for index, key := range p.pKeys {
			ps[key] = p.pReasons[index]
		}
		data[paramsKey] = ps
	}

	return "problem", data
}

func (p *rfc7807) Apply(ctx *Context) {
	if err := ctx.Marshal(p.status, p, true); err != nil {
		ctx.Logs().ERROR().Error(err)
	}

	if len(p.keys) < rfc8707PoolMaxSize && len(p.pKeys) < rfc8707PoolMaxSize {
		rfc7807ProblemPool.Put(p)
	}
}
