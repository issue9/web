// SPDX-License-Identifier: MIT

package module

import (
	"strings"

	"github.com/issue9/logs/v3"
)

// Initializer 定义了初始化功能接口
type Initializer interface {
	// AddInit 添加一个子项
	//
	// 该子项依赖于当前实例，在当前实例执行之后才执行。
	//
	// name 为对该操作的简要描述；
	// f 为该操作实际执行的函数；
	AddInit(name string, f func() error) Initializer
}

type initializer struct {
	name  string
	f     func() error
	inits []*initializer // 子模块
}

func (i *initializer) AddInit(name string, f func() error) Initializer {
	if name == "" {
		panic("参数 name 不能为空")
	}

	if f == nil {
		panic("参数 f 不能为空")
	}

	ii := &initializer{name: name, f: f}
	if i.inits == nil {
		i.inits = []*initializer{ii}
	} else {
		i.inits = append(i.inits, ii)
	}

	return ii
}

func (i *initializer) init(l *logs.Logs, deep int) error {
	indent := strings.Repeat("\t", deep)
	l.Info(indent, i.name)

	if i.f != nil {
		if err := i.f(); err != nil {
			l.Error(indent, err)
			return err
		}
	}

	// 初始化子功能

	deep++
	for _, ii := range i.inits {
		if err := ii.init(l, deep); err != nil {
			return err
		}
	}
	return nil
}
