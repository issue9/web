// SPDX-License-Identifier: MIT

package dep

import "log"

// Item 依赖项
type Item struct {
	id   string   // 唯一 ID
	deps []string // 依赖的其它依赖项 ID

	called    bool
	executors []Executor // 需要执行的函数列表，将按照顺序执行。
}

// Executor 依赖项实际执行的函数对象
type Executor struct {
	Title string
	F     func() error
}

func NewItem(id string, deps []string, executors []Executor) *Item {
	return &Item{
		id:        id,
		deps:      deps,
		executors: executors,
	}
}

func (i *Item) call(info *log.Logger) error {
	const indent = "\t"

	for _, exec := range i.executors {
		info.Printf("%s%s......", indent, exec.Title)
		if err := exec.F(); err != nil {
			info.Printf("%s%s FAIL: %s\n", indent, exec.Title, err.Error())
			return err
		}
		info.Printf("%s%s OK", indent, exec.Title)
	}

	i.called = true

	return nil
}

// Reverse 反转依赖关系
func Reverse(items []*Item) []*Item {
	ret := make([]*Item, 0, len(items))
	for _, item := range items {
		ret = append(ret, &Item{id: item.id, executors: item.executors, deps: []string{}})
	}

	for _, item := range items {
		for _, dep := range item.deps {
			d := findItem(ret, dep)
			d.deps = append(d.deps, item.id)
		}
	}

	return ret
}
