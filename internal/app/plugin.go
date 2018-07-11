// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"plugin"

	"github.com/issue9/web/module"
)

const moduleInitFuncName = "Init"

type moduleInitFunc func(*module.Module)

func (app *App) loadPlugin(path string) (*module.Module, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	m := &module.Module{}

	symbol, err := p.Lookup(moduleInitFuncName)
	if err != nil {
		return nil, err
	}
	init := symbol.(moduleInitFunc)
	init(m)

	return m, nil
}
