// SPDX-License-Identifier: MIT

package module

import "github.com/issue9/web/server"

// Modules 模块管理
type Modules struct {
	modules []*Module
	app     *server.Server
}

// NewModules 声明新的 Modules 对象
func NewModules(app *server.Server, plugin string) (*Modules, error) {
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
