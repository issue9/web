// SPDX-License-Identifier: MIT

package dep

import (
	"fmt"
	"log"
	"sort"

	"github.com/issue9/sliceutil"
)

// Module 模块信息
type Module struct {
	dep  *Dep
	tags map[string]*Tag

	id   string
	desc string
	deps []string
}

// NewModule 声明新的模块
func (d *Dep) NewModule(id, desc string, deps ...string) (*Module, error) {
	if sliceutil.Count(d.modules, func(i int) bool { return d.modules[i].id == id }) > 0 {
		return nil, fmt.Errorf("存在同名的模块 %s", id)
	}

	mod := &Module{
		dep:  d,
		tags: make(map[string]*Tag, 2),

		id:   id,
		desc: desc,
		deps: deps,
	}

	d.modules = append(d.modules, mod)

	return mod, nil
}

// Modules 模块列表
func (dep *Dep) Modules() []*Module { return dep.modules }

// ID 模块的唯一 ID
func (m *Module) ID() string { return m.id }

// Description 对模块的详细描述
func (m *Module) Description() string { return m.desc }

// Deps 模块的依赖信息
func (m *Module) Deps() []string { return m.deps }

// Tags 模块的标签名称列表
func (m *Module) Tags() []string {
	tags := make([]string, 0, len(m.tags))
	for name := range m.tags {
		tags = append(tags, name)
	}
	sort.Strings(tags)
	return tags
}

// Inited 查询指定标签关联的函数是否已经被调用
func (m *Module) Inited(tag string) bool { return m.Tag(tag).Inited() }

// Init 执行关联标签的函数
func (m *Module) Init(l *log.Logger, tag string) (err error) {
	l.Println(m.ID(), "...")

	if err = m.Tag(tag).init(l); err != nil {
		l.Printf("%s [FAIL:%s]\n\n", m.ID(), err.Error())
	} else {
		l.Printf("%s [OK]\n\n", m.ID())
	}

	return err
}
