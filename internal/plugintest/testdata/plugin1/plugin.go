// SPDX-License-Identifier: MIT

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import (
	"fmt"
	"net/http"

	"github.com/issue9/web"
)

// Init 初始化模块
func Init(srv *web.Server) error {
	m := web.NewModule("plugin1", "p1 desc", "plugin2")

	m.AddInit("init1", init1)
	m.AddInit("init2", init2)

	t1 := m.NewTag("install")
	t1.AddInit("title", install1)

	t2 := m.NewTag("v1.0")
	t2.AddInit("title", install2)

	m.Get("/plugin1", func(ctx *web.Context) {
		ctx.Render(http.StatusOK, "plugin1", nil)
	})

	return srv.AddModule(m)
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
