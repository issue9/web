// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import "strconv"

// Init 表示初始化功能的相关数据
type Init struct {
	Title string
	F     func() error
}

// AddInit 添加一个初始化函数
//
// title 该初始化函数的名称。没有则会自动生成一个序号，多个，则取第一个元素。
func (m *Module) AddInit(f func() error, title ...string) *Module {
	if m.Inits == nil {
		m.Inits = make([]*Init, 0, 5)
	}

	t := ""
	if len(title) == 0 {
		t = strconv.Itoa(len(m.Inits))
	} else {
		t = title[0]
	}

	m.Inits = append(m.Inits, &Init{F: f, Title: t})
	return m
}
