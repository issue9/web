// SPDX-License-Identifier: MIT

package serialization

import (
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
		Arg    int    `xml:"arg,attr" json:"arg" yaml:"arg"`
		Format string `xml:"format,attr,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
		Cases  pairs  `xml:"case" json:"cases" yaml:"cases"`
	}

	localeVar struct {
		Name   string `xml:"name,attr" json:"name" yaml:"name"`
		Arg    int    `xml:"arg,attr" json:"arg" yaml:"arg"`
		Format string `xml:"format,attr,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
		Cases  pairs  `xml:"case" json:"cases" yaml:"cases"`
	}

	pairs []pair

	// pair 定义 map[string]string 类型
	//
	// 唯一的功能是为了 xml 能支持 map。
	pair struct {
		Cond  interface{}
		Value interface{}
	}

	entry struct {
		XMLName struct{} `xml:"key"`
		Cond    string   `xml:"cond,attr"`
		Value   string   `xml:",chardata"`
	}
)

func buildSelectf(arg int, format string, cases []pair) catalog.Message {
	c := make([]interface{}, 0, len(cases)*2)
	for _, v := range cases {
		c = append(c, v.Cond, v.Value)
	}
	return plural.Selectf(arg, format, c...)
}

func (v *localeVar) message() catalog.Message {
	msg := buildSelectf(v.Arg, v.Format, v.Cases)
	return catalog.Var(v.Name, msg)
}

func NewLocale(b *catalog.Builder, f *Files) *Locale {
	return &Locale{b: b, f: f}
}

func (b *TagBuilder) SetString(key, msg string) error {
	return b.builder.SetString(b.tag, key, msg)
}

func (b *TagBuilder) Set(key string, msg ...catalog.Message) error {
	return b.builder.Set(b.tag, key, msg...)
}

func (b *TagBuilder) SetMacro(key string, msg ...catalog.Message) error {
	return b.builder.SetMacro(b.tag, key, msg...)
}

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
				msgs = append(msgs, v.message())
			}
			msgs = append(msgs, catalog.String(msg.Message.Msg))
			err = tag.Set(msg.Key, msgs...)
		case msg.Message.Select != nil:
			s := msg.Message.Select
			err = tag.Set(msg.Key, buildSelectf(s.Arg, s.Format, s.Cases))
		case msg.Message.Msg != "":
			err = tag.SetString(msg.Key, msg.Message.Msg)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (e *pairs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	kv := yaml.MapSlice{}
	if err := unmarshal(&kv); err != nil {
		return err
	}

	*e = make(pairs, 0, len(kv))

	// make sure to dereference before assignment,
	// otherwise only the local variable will be overwritten
	// and not the value the pointer actually points to
	for _, item := range kv {
		key := item.Key
		val := item.Value
		*e = append(*e, pair{Cond: key, Value: val})
	}

	return nil
}
