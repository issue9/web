// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"time"

	"github.com/issue9/web/config"
	"github.com/issue9/web/internal/filesystem"
)

// Map 定义 map[string]string 类型
//
// 唯一的功能是为了 xml 能支持 map。
type Map map[string]string

type entry struct {
	XMLName struct{} `xml:"key"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// Debug 调试信息的配置
type Debug struct {
	Pprof string `yaml:"pprof,omitempty" json:"pprof,omitempty" xml:"pprof,omitempty"`
	Vars  string `yaml:"vars,omitempty" json:"vars,omitempty" xml:"vars,omitempty"`
}

// Duration 封装 time.Duration，实现 JSON、XML 和 YAML 的解析
type Duration time.Duration

// Certificate 证书管理
type Certificate struct {
	Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
}

// MarshalXML implement xml.Marshaler
func (p Map) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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

// UnmarshalXML implement xml.Unmarshaler
func (p *Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = Map{}

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

func (cert *Certificate) sanitize() *config.FieldError {
	if !filesystem.Exists(cert.Cert) {
		return &config.FieldError{Field: "cert", Message: "文件不存在"}
	}

	if !filesystem.Exists(cert.Key) {
		return &config.FieldError{Field: "key", Message: "文件不存在"}
	}

	return nil
}
