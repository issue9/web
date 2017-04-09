// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package web 一个微型的 API 框架，包含了以下内容：
//
//
// 路由
//
// 路由源自 https://github.com/issue9/mux 包，包含了正则匹配路由以及路由参数等功能，
// 通过 Router() 函数对外公开相关接口，具体文档可直接参才 mux 包文档。
//
//
// 日志
//
// 日志采用 https://github.com/issue9/logs 包，在 Init() 函数被调用之后，其功能即为可以状态。
//
//
// envelope 模式：
//
// TODO
//
//
// 错误处理
//
// TODO
package web

// Version 当前框架的版本
const Version = "0.9.5+20170409"
