// SPDX-License-Identifier: MIT

package webconfig

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"time"
)

type pairs map[string]string

type entry struct {
	XMLName struct{} `xml:"key"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// Duration 封装 time.Duration，实现 JSON、XML 和 YAML 的解析
type Duration time.Duration

func (p pairs) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(p) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for k, v := range p {
		if err := e.Encode(entry{Name: k, Value: v}); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (p *pairs) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = pairs{}

	for {
		e := &entry{}
		if err := d.Decode(e); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		(*p)[e.Name] = e.Value
	}

	return nil
}

// Duration 转换成 time.Duration
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// MarshalJSON json.Marshaler 接口
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON json.Unmarshaler 接口
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(tmp)
	return nil
}

// MarshalYAML yaml.Marshaler 接口
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// UnmarshalYAML yaml.Unmarshaler 接口
func (d *Duration) UnmarshalYAML(u func(interface{}) error) error {
	var dur time.Duration
	if err := u(&dur); err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

// MarshalXML xml.Marshaler 接口
func (d Duration) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(time.Duration(d).String(), start)
}

// UnmarshalXML xml.Unmarshaler 接口
func (d *Duration) UnmarshalXML(de *xml.Decoder, start xml.StartElement) error {
	var str string
	if err := de.DecodeElement(&str, &start); err != nil && err != io.EOF {
		return err
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(dur)

	return nil
}
