// SPDX-License-Identifier: MIT

package dep

import "log"

// Tag 模块下对执行函数的分类
type Tag struct {
	inited    bool
	executors []executor // 保证按添加顺序执行
}

type executor struct {
	title string
	f     func() error
}

// Tag 返回指定名称的 Tag 实例
//
// 如果不存在则会创建。
func (m *Module) Tag(e string) *Tag {
	ev, found := m.tags[e]
	if !found {
		ev = &Tag{executors: make([]executor, 0, 5)}
		m.tags[e] = ev
	}
	return ev
}

// On 注册指执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (e *Tag) On(title string, f func() error) *Tag {
	e.executors = append(e.executors, executor{title: title, f: f})
	return e
}

func (e *Tag) init(l *log.Logger) error {
	const indent = "\t"

	if e.Inited() {
		return nil
	}

	for _, exec := range e.executors {
		l.Printf("%s%s......", indent, exec.title)
		if err := exec.f(); err != nil {
			l.Printf("%s%s FAIL: %s\n", indent, exec.title, err.Error())
			return err
		}
		l.Printf("%s%s OK", indent, exec.title)
	}

	e.inited = true
	return nil
}

// Inited 当前标签关联的函数是否已经执行过
func (e *Tag) Inited() bool { return e.inited }
