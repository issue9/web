// SPDX-License-Identifier: MIT

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import (
	"fmt"
	"net/http"

	"github.com/issue9/mux/v5/group"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

// InitModule 返回模块信息
func InitModule(s *web.Server) error {
	m, err := s.NewModule("plugin1", "v1", "p1 desc", "plugin2")
	if err != nil {
		return err
	}

	t := m.Tag("default")

	t.On("init1", init1)
	t.On("init2", init2)

	t1 := m.Tag("install")
	t1.On("title", install1)

	t2 := m.Tag("v1.0")
	t2.On("title", install2)

	t.On("init router", func() error {
		r, err := s.NewRouter("p1", "https://example.com", &group.Hosts{})
		if err != nil {
			return err
		}

		r.Get("/plugin1", func(ctx *web.Context) server.Responser {
			return server.Object(http.StatusOK, "plugin1", nil)
		})
		return nil
	})

	return nil
}

func init1() error {
	fmt.Println("plugin1 init1")
	return nil
}

func init2() error {
	fmt.Println("plugin1 init2")
	return nil
}

func install1() error {
	fmt.Println("plugin1 install1")
	return nil
}

func install2() error {
	fmt.Println("plugin1 install2")
	return nil
}

func main() {}
