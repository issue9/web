// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package web 一个微型的 RESTful API 框架。
//
//
// 配置文件
//
// 配置文件的映射对象在 internal/config 包中，其中有各个字段的详细说明。
// 用户如果需要添加一些自定义的配置项，需要自行再添加其它名称的配置文件，
// 文件地址最好通过 web.File 来获取，这样可以和框架本身的配置文件存在同一目录下。
//
// 框架了除了本身的 web.yaml 配置文件之外，还有 logs.xml，用于定制日志的相关内容。
// 具体的日志相关信息，可以访问 https://github.com/issue9/logs 包。
//
//
// 字符集和媒体类型
//
// encoding 包通过 AddMarshal 和 AddUnmarshal 给用户提供相关功能。
//
// 当然用户也可以直接构建一个 context.Context 对象来生成一个一次性的对象。
//
//
// 返回结果
//
// 框架内置了一个 result 包，用以统一向用户返回的错误信息，这是一个可选的包，
// 如果要使用，需要在 web.Run() 之前调用 result.NewMessages() 注册相关的错误代码。
//
//
// 模块
//
// 用户可以把功能相对独立的内容当作一个模块进行封装。框架本身提供了 web.NewModule()
// 对模块进行了依赖管理。用户可以在 web.NewModule() 返回对象中，
// 对模块进行初始化和路由项的添加。所有模块会在 web.Run() 中进行初始化。
package web
