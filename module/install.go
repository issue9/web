// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

type message string

// NewMessage 声明一条 message 类型的错误信息
//
// 返回内容并不是一个真正的错误，则是在某些时候需要在安装完成之后，
// 反馈一些文字信息，则需要用此函数进行包装。
//
// Deprecated: 采用 NewTag 代替
func NewMessage(msg string) error {
	return message(msg)
}

func (msg message) Error() string {
	return string(msg)
}

// NewVersion 为当前模块添加某一版本号下的安装脚本。
//
// Deprecated: 采用 NewTag 代替
func (m *Module) NewVersion(version string) *Module {
	return m.NewTag(version)
}
