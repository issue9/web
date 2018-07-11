// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/issue9/web/module"
)

const moduleInitFuncName = "Init"

func (app *App) loadPlugins() error {
	dir := app.config.PluginDir
	if dir == "" {
		return nil
	}

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(fi.Name())) != ".so" {
			return nil
		}

		m, err := app.loadPlugin(filepath.Join(dir, path))
		if err != nil {
			return err
		}

		app.modules = append(app.modules, m)
		return nil
	})
}

func (app *App) loadPlugin(path string) (*module.Module, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	symbol, err := p.Lookup(moduleInitFuncName)
	if err != nil {
		return nil, err
	}
	init := symbol.(func(*module.Module))

	m := module.New(app.router, "plugin", "plugin desc")
	init(m)

	return m, nil
}
