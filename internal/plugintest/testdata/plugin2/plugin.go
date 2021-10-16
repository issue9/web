// SPDX-License-Identifier: MIT

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import (
	"fmt"

	"github.com/issue9/web"
)

// InitModule 返回模块信息
func InitModule(s *web.Server) error {
	m := s.NewModule("plugin2", "v1", web.Phrase("p2 desc"))

	t := m.Action("default")

	t.AddInit("init1", init1)
	t.AddInit("init2", init2)

	t1 := m.Action("install")
	t1.AddInit("title", install1)

	t2 := m.Action("v1.0")
	t2.AddInit("title", install2)

	return nil
}

func init1() error {
	fmt.Println("plugin2 init1")
	return nil
}

func init2() error {
	fmt.Println("plugin2 init2")
	return nil
}

func install1() error {
	fmt.Println("plugin2 install1")
	return nil
}

func install2() error {
	fmt.Println("plugin2 install2")
	return nil
}

func main() {}
