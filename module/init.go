// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

// 表示初始化功能的相关数据
type initialization struct {
	title string
	f     func() error
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。没有则会自动生成一个序号，多个，则取第一个元素。
func (t *Tag) AddInit(f func() error, title string) *Tag {
	if t.inits == nil {
		t.inits = make([]*initialization, 0, 5)
	}

	t.inits = append(t.inits, &initialization{f: f, title: title})
	return t
}
