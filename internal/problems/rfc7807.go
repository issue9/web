// SPDX-License-Identifier: MIT

package problems

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

type ctx interface {
	Render(int, any, bool)
}

// RFC7807 server.Problem 接口的 rfc7807 实现
//
// C 表示实现的 server.Context 类型，
// 因为此类型最终会被 server 包引用，为了拆分代码，用泛型代替。
type RFC7807[C ctx] struct {
	pool *RFC7807Pool[C]

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

type RFC7807Pool[C ctx] struct {
	pool *sync.Pool
}

func NewRFC7807Pool[C ctx]() *RFC7807Pool[C] {
	return &RFC7807Pool[C]{
		pool: &sync.Pool{New: func() any {
			return &RFC7807[C]{
				keys: []string{"type", "title", "status", "detail"},
				vals: make([]any, 4),
			}
		}},
	}
}

// New 获取 RFC7807 对象
func (pool *RFC7807Pool[C]) New(id string, status int, title, detail string) *RFC7807[C] {
	p := pool.pool.Get().(*RFC7807[C])
	p.pool = pool
	p.status = status

	p.keys = p.keys[:4]
	p.vals = p.vals[:4]
	// keys 前三个元素为固定值
	p.vals[0] = id
	p.vals[1] = title
	p.vals[2] = status
	p.vals[3] = detail

	p.pKeys = p.pKeys[:0]
	p.pReasons = p.pReasons[:0]

	return p
}

func (p *RFC7807[C]) MarshalJSON() ([]byte, error) {
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

func (p *RFC7807[C]) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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

func (p *RFC7807[C]) MarshalForm() ([]byte, error) {
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

func (p *RFC7807[C]) MarshalHTML() (string, any) {
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

func (p *RFC7807[C]) Apply(ctx C) {
	ctx.Render(p.status, p, true)
	if len(p.keys) < rfc8707PoolMaxSize && len(p.pKeys) < rfc8707PoolMaxSize {
		p.pool.pool.Put(p)
	}
}

func (p *RFC7807[C]) AddParam(name string, reason string) {
	if _, found := sliceutil.At(p.pKeys, func(pp string) bool { return pp == name }); found {
		panic("已经存在")
	}
	p.pKeys = append(p.pKeys, name)
	p.pReasons = append(p.pReasons, reason)
}

func (p *RFC7807[C]) With(key string, val any) {
	if sliceutil.Exists(p.keys, func(e string) bool { return e == key }) || key == paramsKey {
		panic("存在同名的参数")
	}
	p.keys = append(p.keys, key)
	p.vals = append(p.vals, val)
}
