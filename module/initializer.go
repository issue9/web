// SPDX-License-Identifier: MIT

package module

import (
	"strings"

	"github.com/issue9/logs/v2"
)

// Initializer 定义了初始化功能接口
type Initializer interface {
	// AddInit 添加一个初始化函数
	//
	// name 为该操作的名称；
	// f 为该操作实际执行的函数；
	// 如果有后续的操作是依赖当前功能的，可以通过返回的实例调用其 AddInit 方法。
	AddInit(name string, f func() error) Initializer
}

type initializer struct {
	name  string
	f     func() error
	inits []*initializer // 子模块
}

// AddInit 添加一个子项
//
// 该子项依赖于当前实例，在当前实例执行之后才执行。
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
