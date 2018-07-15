// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Load 加载指定的文件
func Load(path string) (*Web, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Web{}
	if err = yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}

	if err = conf.sanitize(); err != nil {
		return nil, err
	}

	return conf, nil
}
