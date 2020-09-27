// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
//
//
// 配置文件
//
// 配置文件的映射对象在 web.go 文件，其中有各个字段的详细说明。
// 用户如果需要添加一些自定义的配置项，需要自行再添加其它名称的配置文件，
// 并通过 config.LoadFile 加载相应的配置文件。
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
// 媒体类型
//
// 默认情况下，框架不会处理任何的 mimetype 类型的数据。需要用户通过 Web
// 对象中的 Marshalers 和 Unmarshalers 字段指定相应的解码和编码函数，
// context/mimetype/ 之下也包含部分已经实现的编解码函数。
//
//
// 返回结果
//
// context 包下的 Result 表示在出错时的输出内容。在使用前，用户需要在
// web.Results 添加各类错误代码。
//
//
// 模块
//
// 用户可以把功能相对独立的内容当作一个模块进行封装。框架本身提供了 web.MODServer
// 对模块进行了依赖管理。用户可以在 web.NewModule() 返回对象中，
// 对模块进行初始化和路由项的添加。所有模块会在 web.Init() 中进行初始化。
package web
