// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"strconv"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
)

type compressConfig struct {
	// Type content-type 的值
	//
	// 可以带通配符，比如 text/* 表示所有 text/ 开头的 content-type 都采用此压缩方法。
	Types []string `json:"types" xml:"type" yaml:"types" toml:"types"`

	// IDs 压缩方法的 ID 列表
	//
	// 这些 ID 值必须是由 [RegisterCompressor] 注册的，否则无效，默认情况下支持以下类型：
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
	ID string `json:"id" xml:"id,attr" yaml:"id" toml:"id"`
}

type mediaTypeConfig struct {
	// 实际采用的解码方法
	//
	// 由 [RegisterMediaType] 注册而来。默认可用为：
	//
	//  - xml
	//  - cbor
	//  - json
	//  - form
	//  - gob
	//  - yaml
	//  - nop  没有具体实现的方法，对于上传等需要自行处理的情况可以指定此值。
	Target string `json:"target" yaml:"target" xml:"target,attr" toml:"target"`
}

func (conf *configOf[T]) buildNegotiation(f *filter.Filter) {
	if len(conf.Compressors) == 0 && len(conf.MediaTypes) == 0 {
		return
	}

	c := web.NewNegotiation(conf.languageTag, conf.config)

	for index, e := range conf.Compressors {
		if enc, found := compressorFactory.get(e.ID); !found {
			f.AddReason("compresses["+strconv.Itoa(index)+"].id", locales.NotFound)
		} else {
			c.AddCompressor(enc, e.Types...)
		}
	}

	conf.sanitizeMediaTypes(c, f)
}

func (conf *configOf[T]) sanitizeMediaTypes(c *web.Negotiation, f *filter.Filter) {
	for index, item := range conf.MediaTypes {
		m, found := mediaTypeFactory.get(item.Target)
		if !found {
			f.AddReason("mediaTypes["+strconv.Itoa(index)+"].target", locales.NotFound)
			return
		}

		c.AddMediaType(m)
	}

	conf.negotiation = c
}
