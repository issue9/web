// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"strconv"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
)

type compressConfig struct {
	// Type content-type 的值
	//
	// 可以带通配符，比如 text/* 表示所有 text/ 开头的 content-type 都采用此压缩方法。
	Types []string `json:"types" xml:"type" yaml:"types"`

	// IDs 压缩方法的 ID 列表
	//
	// 这些 ID 值必须是由 [RegisterCompress] 注册的，否则无效，默认情况下支持以下类型：
	//  - deflate-default
	//  - deflate-best-compression
	//  - deflate-best-speed
	//  - gzip-default
	//  - gzip-best-compression
	//  - gzip-best-speed
	//  - compress-lsb-8
	//  - compress-msb-8
	//  - br-default
	//  - br-best-compression
	//  - br-best-speed
	//  - zstd-default
	ID string `json:"id" xml:"id,attr" yaml:"id"`
}

type mimetypeConfig struct {
	// 编码名称
	//
	// 比如 application/xml 等
	Type string `json:"type" yaml:"type" xml:"type,attr"`

	// 返回错误代码是的 mimetype
	//
	// 比如正常情况下如果是 application/json，那么此值可以是 application/problem+json。
	// 如果为空，表示与 Type 相同。
	Problem string `json:"problem,omitempty" yaml:"problem,omitempty" xml:"problem,attr,omitempty"`

	// 实际采用的解码方法
	//
	// 由 [RegisterMimetype] 注册而来。默认可用为：
	//
	//  - xml
	//  - cbor
	//  - json
	//  - form
	//  - html
	//  - gob
	//  - yaml
	//  - nop  没有具体实现的方法，对于上传等需要自行处理的情况可以指定此值。
	Target string `json:"target" yaml:"target" xml:"target,attr"`
}

type mimetype struct {
	marshal   web.MarshalFunc
	unmarshal web.UnmarshalFunc
}

func (conf *configOf[T]) buildCodec() *web.FieldError {
	if len(conf.Compressors) == 0 && len(conf.Mimetypes) == 0 {
		return nil
	}

	c := web.NewCodec()

	for index, e := range conf.Compressors {
		enc, found := compressorFactory.get(e.ID)
		if !found {
			field := "compresses[" + strconv.Itoa(index) + "].id"
			return web.NewFieldError(field, locales.ErrNotFound())
		}

		c.AddCompressor(enc, e.Types...)
	}

	return conf.sanitizeMimetypes(c)
}

func (conf *configOf[T]) sanitizeMimetypes(c *web.Codec) *web.FieldError {
	if indexes := sliceutil.Dup(conf.Mimetypes, func(i, j *mimetypeConfig) bool { return i.Type == j.Type }); len(indexes) > 0 {
		value := conf.Mimetypes[indexes[1]].Type
		err := web.NewFieldError("mimetypes["+strconv.Itoa(indexes[1])+"].type", locales.DuplicateValue)
		err.Value = value
		return err
	}

	for index, item := range conf.Mimetypes {
		m, found := mimetypesFactory.get(item.Target)
		if !found {
			return web.NewFieldError("mimetypes["+strconv.Itoa(index)+"].target", locales.ErrNotFound())
		}

		c.AddMimetype(item.Type, m.marshal, m.unmarshal, item.Problem)
	}

	conf.codec = c
	return nil
}
