// SPDX-License-Identifier: MIT

package dep

import "log"

type Item struct {
	ID   string
	Deps []string

	called    bool
	Executors []Executor
}

func (i *Item) call(info *log.Logger) error {
	const indent = "\t"

	for _, exec := range i.Executors {
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

type Executor struct {
	Title string
	F     func() error
}

// Reverse 反转依赖关系
func Reverse(items []*Item) []*Item {
	ret := make([]*Item, 0, len(items))
	for _, item := range items {
		ret = append(ret, &Item{ID: item.ID, Executors: item.Executors, Deps: []string{}})
	}

	for _, item := range items {
		for _, dep := range item.Deps {
			d := findItem(ret, dep)
			d.Deps = append(d.Deps, item.ID)
		}
	}

	return ret
}
