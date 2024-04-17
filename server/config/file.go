// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"strconv"

	"github.com/issue9/config"

	"github.com/issue9/web"
)

type fileSerializer struct {
	exts      []string // 支持的扩展名
	marshal   config.MarshalFunc
	unmarshal config.UnmarshalFunc
}

func (conf *configOf[T]) buildConfig() *web.FieldError {
	ss := make(config.Serializer, len(fileSerializerFactory.items))
	for i, name := range conf.FileSerializers {
		s, found := fileSerializerFactory.get(name)
		if !found {
			return web.NewFieldError("fileSerializers["+strconv.Itoa(i)+"]", web.NewLocaleError("not found serialization function for %s", name))
		}
		ss.Add(s.marshal, s.unmarshal, s.exts...)
	}

	c, err := config.BuildDir(ss, conf.dir)
	if err != nil {
		return web.NewFieldError("", err) // 应该是与目录相关的错误引起的
	}

	conf.config = c
	return nil
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
	return conf, nil
}

func buildSerializerFromFactory() config.Serializer {
	s := make(config.Serializer, len(fileSerializerFactory.items))
	for _, item := range fileSerializerFactory.items {
		s.Add(item.marshal, item.unmarshal, item.exts...)
	}
	return s
}
