// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"sort"
	"sync"
)

// Apps 多个 App 实例的集合。可以实现多站点服务。
// 相关功能可以在 https://github.com/issue9/web/issues/5 进行跟踪。
type Apps struct {
	apps []*App
}

// NewApps 新的 Apps 实例。
func NewApps(app ...*App) *Apps {
	if len(app) == 0 {
		panic("参数不能为空")
	}

	ports := make([]int, 0, len(app))
	for _, a := range app {
		ports = append(ports, a.webConfig.Port)
	}
	sort.Ints(ports)
	for i := 1; i < len(ports); i++ {
		if ports[i-1] == ports[i] {
			panic("存在相同的端口号")
		}
	}

	return &Apps{
		apps: app,
	}
}

// Serve 运行服务
//
// 如果所有服务都结束，返回 http.ErrServeClosed 错误
func (apps *Apps) Serve() error {
	wg := &sync.WaitGroup{}

	for _, app := range apps.apps {
		wg.Add(1)
		go func(app *App) {
			app.Serve()
			wg.Done()
		}(app)
	}

	wg.Wait()
	return http.ErrServerClosed
}
