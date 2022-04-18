// SPDX-License-Identifier: MIT

package app

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"

	"github.com/issue9/web/serialization"
)

var encodingFactory = map[string]serialization.EncodingWriterFunc{}

type encodingsConfig struct {
	// 忽略对此类 mimetype 内容的压缩
	//
	// 当用户请求的 accept 报头与此列表相匹配时，将不会对此请求的内容进行压缩。
	// 可以有通配符，比如 image/* 表示任意 image/ 开头的内容。
	// 默认为空。
	Ignores []string `json:"ignores,omitempty" xml:"ignore,omitempty" yaml:"ignores,omitempty"`

	// 压缩方法
	//
	// 键名为压缩名，比如 gzip，flate 等，键值为生成对象的方法。
	// 若为空，则不支持压缩功能。
	Encodings []*encodingConfig `xml:"encoding,omitempty" yaml:"encodings,omitempty" json:"encodings,omitempty"`
}

type encodingConfig struct {
	Name     string `json:"name" xml:"name,attr" yaml:"name"`
	Encoding string `json:"encoding" xml:"encoding,attr" yaml:"encoding"`
}

func (conf *encodingsConfig) build(l logs.Logger) (*serialization.Encodings, *ConfigError) {
	if conf == nil {
		return serialization.NewEncodings(l), nil
	}

	es := make(map[string]serialization.EncodingWriterFunc)
	for _, e := range conf.Encodings {
		f, found := encodingFactory[e.Encoding]
		if !found {
			return nil, &ConfigError{Field: "encodings[" + e.Encoding + "]", Message: localeutil.Error("%s not found", e.Encoding)}
		}
		es[e.Name] = f
	}

	encoding := serialization.NewEncodings(l, conf.Ignores...)
	encoding.Add(es)
	return encoding, nil
}

func RegisterEncoding(f serialization.EncodingWriterFunc, name string) {
	encodingFactory[name] = f
}

func init() {
	RegisterEncoding(serialization.DeflateWriter, "deflate")
	RegisterEncoding(serialization.BrotliWriter, "brotli")
	RegisterEncoding(serialization.GZipWriter, "gzip")
}
