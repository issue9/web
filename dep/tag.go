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
func (m *Module) Tag(t string) *Tag {
	ev, found := m.tags[t]
	if !found {
		ev = &Tag{executors: make([]executor, 0, 5)}
		m.tags[t] = ev
	}
	return ev
}

// AddInit 注册指执行函数
//
// NOTE: 按添加顺序执行各个函数。
func (t *Tag) AddInit(title string, f func() error) *Tag {
	t.executors = append(t.executors, executor{title: title, f: f})
	return t
}

func (t *Tag) init(l *log.Logger) error {
	const indent = "\t"

	if t.Inited() {
		return nil
	}

	for _, exec := range t.executors {
		l.Printf("%s%s......", indent, exec.title)
		if err := exec.f(); err != nil {
			l.Printf("%s%s FAIL: %s\n", indent, exec.title, err.Error())
			return err
		}
		l.Printf("%s%s OK", indent, exec.title)
	}

	t.inited = true
	return nil
}

// Inited 当前标签关联的函数是否已经执行过
func (e *Tag) Inited() bool { return e.inited }
