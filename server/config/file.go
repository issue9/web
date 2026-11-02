// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"strconv"

	"github.com/issue9/config"
	"github.com/issue9/web/filter"

	"github.com/issue9/web"
)

type fileSerializer struct {
	exts      []string // 支持的扩展名
	marshal   config.MarshalFunc
	unmarshal config.UnmarshalFunc
}

func (conf *configOf[T]) buildConfig(f *filter.Filter) {
	ss := make(config.Serializer, len(fileSerializerFactory.items))
	for i, name := range conf.FileSerializers {
		s, found := fileSerializerFactory.get(name)
		if !found {
			f.AddReason("fileSerializers["+strconv.Itoa(i)+"]", web.Phrase("not found serialization function for %s", name))
			return
		}
		ss.Add(s.marshal, s.unmarshal, s.exts...)
	}

	c, err := config.BuildDir(ss, conf.dir)
	if err != nil {
		f.AddError("dir", err)
	}

	conf.config = c
}

func loadConfigOf[T comparable](configDir, name string) (*configOf[T], error) {
	c, err := config.BuildDir(buildSerializerFromFactory(), configDir)
	if err != nil {
		return nil, err
	}

	conf := &configOf[T]{dir: configDir}
	if err := c.Load(name, conf); err != nil {
		return nil, err
	}

	if ferr := conf.filter(name); ferr != nil {
		return nil, ferr
	}
	return conf, nil
}

func buildSerializerFromFactory() config.Serializer {
	s := make(config.Serializer, len(fileSerializerFactory.items))
	for _, item := range fileSerializerFactory.items {
		s.Add(item.marshal, item.unmarshal, item.exts...)
	}
	return s
}
