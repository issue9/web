// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package web 一个微型的 RESTful API 框架。
//
//
// 配置文件
//
// 配置文件的映射对象在 internal/webconfig 包中，其中有各个字段的详细说明。
// 用户如果需要添加一些自定义的配置项，需要自行再添加其它名称的配置文件，
// 文件地址最好通过 web.File 来获取，这样可以和框架本身的配置文件存在同一目录下。
//
// 框架了除了本身的 web.yaml 配置文件之外，还有 logs.xml，用于定制日志的相关内容。
// 具体的日志相关信息，可以访问 https://github.com/issue9/logs 包。
//
//
// 字符集
//
// 字符集用户无须任何操作，会自动根据 `Content-Type` 中的 charset
// 属性自动解析其字符集，
// 输出时，也会根据 `Accept-Charset` 报头内容，作自动转换之后再输出。
// 以下字符集都被支持：
// https://www.iana.org/assignments/character-sets/character-sets.xhtml
//
//
// 字符集和媒体类型
//
// 默认情况下，框架不会处理任何的 mimetype 类型的数据。需要用户通过
// Mimetypes().AddMarshals() 和 Mimetypes().AddUnmarshals() 添加相关的处理函数。
// 添加方式如下：
//  Mimetypes().AddMarshals(map[string]mimetype.MarshalFunc{
//      "application/json": json.Marshal,
//  })
//  Mimetypes().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
//      "application/json": json.Unmarshal,
//  })
// 之后，通过 web.NewContext() 获得的 context 对象，会根据用户的
// Accept 和 Content-Type 自动使用相应的解析和输出格式。
//
// 当然用户也可以直接构建一个 context.Context 对象来生成一个一次性的对象。
//
//
// 返回结果
//
// context 包下的 Result 表示在出错时的输出内容。在使用前，用户需要调用 web.NewMessages()
// 添加各类错误代码。
//
//
// 模块
//
// 用户可以把功能相对独立的内容当作一个模块进行封装。框架本身提供了 web.NewModule()
// 对模块进行了依赖管理。用户可以在 web.NewModule() 返回对象中，
// 对模块进行初始化和路由项的添加。所有模块会在 web.Run() 中进行初始化。
package web
