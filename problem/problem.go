// SPDX-License-Identifier: MIT

// Package problem 对客户端输出非正常信息的处理
package problem

import "github.com/issue9/validation"

// Problem 错误信息对象需要实现的接口
//
// 字段参考了 [RFC7807]，但是没有固定具体的呈现方式，
// 用户可以自定义具体的渲染方法。可以使用 [RFC7807Problem]。
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
type Problem interface {
	GetType() string
	SetType(string)

	GetTitle() string
	SetTitle(string)

	GetDetail() string
	SetDetail(string)

	GetStatus() int
	SetStatus(int)

	GetInstance() string
	SetInstance(string)

	// Destroy 销毁当前对象
	//
	// 如果当前对象采用了类似 sync.Pool 的技术对内容进行了保留，
	// 那么可以在此方法中调用 [sync.Pool.Put] 方法返回给对象池。
	// 否则的话可以实现为一个空方法即可。
	Destroy()
}

type FieldErrs = validation.LocaleMessages
