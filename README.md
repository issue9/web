web
[![Test](https://github.com/issue9/web/workflows/Test/badge.svg)](https://github.com/issue9/web/actions?query=workflow%3ATest)
[![Go version](https://img.shields.io/badge/Go-1.14-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/web)](https://goreportcard.com/report/github.com/issue9/web)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/web/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/web)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/web)](https://pkg.go.dev/github.com/issue9/web)
======

web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能。

如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

```go
package main

import "github.com/issue9/web"

// main.go
func main() {
    w, _ := web.Classic("./appconfig/logs.xml", "./appconfig/web.yaml")

    // 注册模块信息
    m1.Init()
    m2.Init()

    w.Serve()
}

// modules/m1/module.go
func Init(s *web.Web) {
    s.NewModule("m1", "模块描述信息").
        Get("/admins", getAdmins).
        Get("/groups", getGroups)
}

// modules/m2/module.go
func Init(s *web.Web) {
    s.NewModule("m2", "模块描述信息", "m1").
        Get("/admins", getAdmins).
        Get("/groups", getGroups)
}
```

项目结构
---

这只是推荐的目录结构，但不是必须按照此来。

```text
+----- common 一些公用的包
|
+----- modules 各个模块的代码
|        |
|        +----- module1
|        |
|        +----- module2
|
+----- cmd
|        |
|        +----- main.go
|        |
|        |----- appconfig 配置文存放路径
|                  |
|                  +----- web.yaml 框架本身的配置文件
|                  |
|                  +----- logs.xml 日志配置文件
|
```

模块
---

项目主要代码都在 modules 下的各个模块里，每一个模块需要包含一个初始化函数，
用于向框架注册当前模块的一些主要信息。通过 `web.Web` 注册模块：

```go
package m1

import "github.com/issue9/web"

func Init(s *web.Web) {
    m := s.NewModule("test", "测试模块")

    m.AddInit(func() error {
        // TODO 此处可以添加初始化模块的相关代码
        return nil
    }, "初始化函数描述")

    m.AddService(func(ctx context.Context) error {
        // TODO 此处添加服务代码
    }, "服务描述")
}
```

#### 字符集

字符集用户无须任何操作，会自动根据 `Content-Type` 中的 charset 属性自动解析其字符集，
输出时，也会根据 `Accept-Charset` 报头内容，作自动转换之后再输出。以下字符集都被支持：
<https://www.iana.org/assignments/character-sets/character-sets.xhtml>

#### 媒体类型

默认情况下，框架不会处理任何的 mimetype 类型的数据。需要用户通过
`Config.Marhsalers` 和 `Config.Unmarshalers` 添加相关的处理函数。
添加方式如下：

```go
conf := &web.Config {
    Marshalers: map[string]mimetype.MarhsalFunc{
        "application/json": json.Marshal,
    },
    Unmarshalers: map[string]mimetype.UnmarhsalFunc{
        "application/json": json.Unmarshal,
    },
    // 其它设置项
}

srv := web.New(conf)
srv.Serve()
```

#### 错误处理

框架提供了一种输出错误信息内容的机制，用户只需要实现 Result 接口，即可自定义输出的错误信息格式。
具体实现可参考 context.defaultResult 的实现。

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
