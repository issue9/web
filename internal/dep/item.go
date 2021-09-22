// SPDX-License-Identifier: MIT

package dep

type Item struct {
	ID   string
	Deps []string

	called    bool
	Executors []Executor
}

func (i *Item) call() error {
	for _, e := range i.Executors {
		if err := e.F(); err != nil {
			return err
		}
	}
	i.called = true

	return nil
}

type Executor struct {
	Title string
	F     func() error
}

// 反转依赖关系
func reverseItems([]*Item) {
	// TODO
}
