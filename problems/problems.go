// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package problems 提供对  [RFC7807] 和 [RFC9457] 管理
//
// [RFC7807]: https://www.rfc-editor.org/info/rfc7807
// [RFC9457]: https://www.rfc-editor.org/info/rfc9457
package problems

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/internal/status"
)

type Builder[T any] func(*Item, *localeutil.Printer, *T)

// Problems 管理 Problem 对象
type Problems[T any] struct {
	problems []*Item // 需保证元素的顺序相同
	build    Builder[T]
}

type Item struct {
	// ID 唯一 ID
	//
	// 与 Type 的区别在于 ID 用于查找，需要唯一性。
	// 而 Type 可能是 'about:blank' 或是不同的 URI。
	ID string

	Type          string
	Title, Detail localeutil.Stringer
	Status        int
}

// New 新建 [Problems] 对象
//
// build 提供构建 problem 对象的方法
func New[T any](build Builder[T]) *Problems[T] {
	ps := &Problems[T]{
		problems: make([]*Item, 0, 100),
		build:    build,
	}
	ps.initProblems()
	return ps
}

// Add 添加一组新项
//
// 在 p.Type 不是绝对 URL 的情况下，typePrefix 会自动加在其前部.做为前缀。
//
// NOTE: 已添加的内容无法修改，如果确实有需求，只能通过修改翻译项的方式间接进行修改。
func (ps *Problems[T]) Add(typePrefix string, p ...*Item) *Problems[T] {
	for _, pp := range p {
		if pp.Type != AboutBlank && typePrefix != "" &&
			!strings.HasPrefix(pp.Type, "https://") && !strings.HasPrefix(pp.Type, "http://") {
			pp.Type = typePrefix + pp.Type
		}

		ps.add(pp)
	}
	return ps
}

func (ps *Problems[T]) add(p *Item) *Problems[T] {
	if !status.IsProblemStatus(p.Status) { // 只需验证大于 400 的状态码。
		panic("status 必须是一个有效的状态码")
	}

	if p.ID == "" {
		panic("p.ID 不能为空")
	}

	if ps.Exists(p.ID) {
		panic(fmt.Sprintf("该 ID 值 %s 已经存在", p.ID))
	}

	if p.Type == "" {
		p.Type = AboutBlank
	}
	if p.Type != AboutBlank {
		if slices.ContainsFunc(ps.problems, func(pp *Item) bool { return pp.Type == p.Type }) {
			panic(fmt.Sprintf("该 Type 值 %s 已经存在", p.Type))
		}
	}

	if p.Title == nil {
		panic("title 不能为空")
	}

	ps.problems = append(ps.problems, p)

	return ps
}

// Exists 查看指定 id 是否已经存在
func (ps *Problems[T]) Exists(id string) bool {
	return slices.IndexFunc(ps.problems, func(p *Item) bool { return p.ID == id }) > -1
}

// All 所有已注册的错误代码
func (ps *Problems[T]) All() iter.Seq[*Item] {
	return func(yield func(*Item) bool) {
		for _, p := range ps.problems {
			if !yield(p) {
				break
			}
		}
	}
}

func (ps *Problems[T]) Init(id string, printer *localeutil.Printer, p *T) {
	if i := slices.IndexFunc(ps.problems, func(lp *Item) bool { return lp.ID == id }); i > -1 {
		lp := ps.problems[i]
		ps.build(lp, printer, p)
		return
	}
	panic(fmt.Sprintf("未找到有关 %s 的定义", id)) // 初始化时没有给定相关的定义，所以直接 panic。
}

// IsProblem 这是表示错误的状态码
func IsProblem(s int) bool { return status.IsProblemStatus(s) }
