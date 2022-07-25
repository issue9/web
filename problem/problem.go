// SPDX-License-Identifier: MIT

// Package problem 对客户端输出非正常信息的处理
package problem

import "github.com/issue9/validation"

// Problem 错误信息对象需要实现的接口
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type Problem interface {
	// 获取反馈给客户端的状态码
	Status() int

	// 设置错误的出处
	SetInstance(string)

	// 添加验证错误的信息
	AddParam(name string, reason ...string)

	// Destroy 销毁当前对象
	//
	// 如果当前对象采用了类似 sync.Pool 的技术对内容进行了保留，
	// 那么可以在此方法中调用 [sync.Pool.Put] 方法返回给对象池。
	// 否则的话可以实现为一个空方法即可。
	Destroy()
}

type BuildFunc func(id, title, detail string, status int) Problem

type FieldErrs = validation.LocaleMessages
