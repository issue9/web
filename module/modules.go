// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import "github.com/issue9/web/app"

// Modules 模块管理
type Modules struct {
	modules []*Module
	app     *app.App
}

// NewModules 声明新的 Modules 对象
func NewModules(app *app.App, plugin string) (*Modules, error) {
	ms := &Modules{
		modules: make([]*Module, 0, 100),
		app:     app,
	}

	if plugin != "" {
		if err := ms.loadPlugins(plugin); err != nil {
			return nil, err
		}
	}

	return ms, nil
}
