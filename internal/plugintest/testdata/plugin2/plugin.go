// SPDX-License-Identifier: MIT

// 测试用的 plugin 模块，可以直接运行 go generate 生成 .so 文件
package main

import (
	"fmt"

	"github.com/issue9/web"
)

// InitModule 返回模块信息
func InitModule(s *web.Server) error {
	m, err := s.NewModule("plugin2", "v1", "p2 desc")
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
