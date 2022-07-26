// SPDX-License-Identifier: MIT

// Package problem 对客户端输出非正常信息的处理
package problem

// Problem API 错误信息对象需要实现的接口
//
// 除了当前接口，该对象可能还要实现相应的序列化接口，比如要能被 JSON 解析，
// 就要实现 [json.Marshaler] 接口，或是相应的 struct tag。
//
// 并未规定 [Problem] 实现都输出的字段以及布局，实现者可以根据 [BuildFunc]
// 给定的参数，结合自身需求决定。比如 [RFC7807Builder] 实现了一个简要的
// RFC7807 标准的错误信息对象。
type Problem interface {
	// 获取反馈给客户端的状态码
	Status() int

	// 添加新的输出字段
	//
	// 如果添加的字段名称与现有的字段重名，应当 panic。
	With(key string, val any)

	// 添加验证错误的信息
	AddParam(name string, reason ...string)

	// Destroy 销毁当前对象
	//
	// 如果当前对象采用了类似 sync.Pool 的技术对内容进行了保留，
	// 那么可以在此方法中调用 [sync.Pool.Put] 方法返回给对象池。
	// 否则的话可以实现为一个空方法即可。
	Destroy()
}

// BuildFunc 生成 [Problem] 对象的方法
//
// id 表示当前错误信息的唯一值，这将是一个标准的 URL，指向线上的文档地址；
// title 错误信息的简要描述；
// status 输出的状态码，该值将由 [Problem.Status] 返回；
type BuildFunc func(id, title string, status int) Problem
