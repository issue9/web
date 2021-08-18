// SPDX-License-Identifier: MIT

package serialization

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"
)

type (
	// Locale 提供了从文件中加载本地化信息的方法
	Locale struct {
		b *catalog.Builder
		f *Files
	}

	// TagBuilder 简化 catalog.Builder
	//
	// 为 catalog.Builder.SetString 省略了 tag 参数
	TagBuilder struct {
		builder *catalog.Builder
		tag     language.Tag
	}

	localeMessages struct {
		XMLName  struct{}        `xml:"messages" json:"-" yaml:"-"`
		Language language.Tag    `xml:"language,attr" json:"language" yaml:"language"`
		Messages []localeMessage `xml:"message" json:"messages" yaml:"messages"`
	}

	// localeMessage 单条消息
	localeMessage struct {
		Key     string     `xml:"key" json:"key" yaml:"key"`
		Message localeText `xml:"message" json:"message" yaml:"message"`
	}

	localeText struct {
		Msg    string        `xml:"msg,omitempty" json:"msg,omitempty"  yaml:"msg,omitempty"`
		Select *localeSelect `xml:"select,omitempty" json:"select,omitempty" yaml:"select,omitempty"`
		Vars   []*localeVar  `xml:"var,omitempty" json:"vars,omitempty" yaml:"vars,omitempty"`
	}

	localeSelect struct {
		Arg    int         `xml:"arg,attr" json:"arg" yaml:"arg"`
		Format string      `xml:"format,attr,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
		Cases  localeCases `xml:"case" json:"cases" yaml:"cases"`
	}

	localeVar struct {
		Name   string      `xml:"name,attr" json:"name" yaml:"name"`
		Arg    int         `xml:"arg,attr" json:"arg" yaml:"arg"`
		Format string      `xml:"format,attr,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
		Cases  localeCases `xml:"case" json:"cases" yaml:"cases"`
	}

	localeCases []interface{}

	entry struct {
		XMLName struct{} `xml:"case"`
		Cond    string   `xml:"cond,attr"`
		Value   string   `xml:",chardata"`
	}
)

// NewLocale 返回 Locale 实例
//
// f 表示用于加载本地化文件的序列化方法，根据文件扩展名在 f 中查找相应的序列化方法；
// 加载后的内容被应用在 b 之上。
func NewLocale(b *catalog.Builder, f *Files) *Locale { return &Locale{b: b, f: f} }

func (b *TagBuilder) SetString(key, msg string) error {
	return b.builder.SetString(b.tag, key, msg)
}

func (b *TagBuilder) Set(key string, msg ...catalog.Message) error {
	return b.builder.Set(b.tag, key, msg...)
}

func (b *TagBuilder) SetMacro(key string, msg ...catalog.Message) error {
	return b.builder.SetMacro(b.tag, key, msg...)
}

// Files 返回用于序列化文件的实例
func (l *Locale) Files() *Files { return l.f }

// Builder 返回本地化操作的相关接口
func (l *Locale) Builder() *catalog.Builder { return l.b }

func (l *Locale) TagBuilder(tag language.Tag) *TagBuilder {
	return &TagBuilder{
		builder: l.Builder(),
		tag:     tag,
	}
}

func (l *Locale) Printer(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(l.Builder()))
}

// LoadFile 从文件中加载本地化内容
func (l *Locale) LoadFile(glob string) error {
	matchs, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, f := range matchs {
		m := &localeMessages{}
		if err := l.f.Load(f, m); err != nil {
			return err
		}

		if err := l.set(m); err != nil {
			return err
		}
	}

	return nil
}

// LoadFileFS 从文件中加载本地化内容
func (l *Locale) LoadFileFS(fsys fs.FS, glob string) error {
	matchs, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	for _, f := range matchs {
		m := &localeMessages{}
		if err := l.f.LoadFS(fsys, f, m); err != nil {
			return err
		}

		if err := l.set(m); err != nil {
			return err
		}
	}

	return nil
}

func (l *Locale) set(m *localeMessages) (err error) {
	tag := l.TagBuilder(m.Language)

	for _, msg := range m.Messages {
		switch {
		case msg.Message.Vars != nil:
			vars := msg.Message.Vars
			msgs := make([]catalog.Message, 0, len(vars))
			for _, v := range vars {
				mm := catalog.Var(v.Name, plural.Selectf(v.Arg, v.Format, v.Cases...))
				msgs = append(msgs, mm)
			}
			msgs = append(msgs, catalog.String(msg.Message.Msg))
			err = tag.Set(msg.Key, msgs...)
		case msg.Message.Select != nil:
			s := msg.Message.Select
			err = tag.Set(msg.Key, plural.Selectf(s.Arg, s.Format, s.Cases...))
		case msg.Message.Msg != "":
			err = tag.SetString(msg.Key, msg.Message.Msg)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalXML implement xml.Unmarshaler
func (c *localeCases) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		e := &entry{}
		if err := d.DecodeElement(e, &start); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}

		*c = append(*c, e.Cond, e.Value)
	}
}

func (c *localeCases) UnmarshalYAML(unmarshal func(interface{}) error) error {
	kv := yaml.MapSlice{}
	if err := unmarshal(&kv); err != nil {
		return err
	}

	*c = make(localeCases, 0, len(kv))
	for _, item := range kv {
		*c = append(*c, item.Key, item.Value)
	}

	return nil
}

func (c *localeCases) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	for {
		t, err := d.Token()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}

		if t == json.Delim('{') || t == json.Delim('}') {
			continue
		}

		*c = append(*c, t)
	}
}
