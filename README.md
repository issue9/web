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

模块
---

在 web 中项目以模块进行划分。每个模块返回一个 *web.Module 实例，
向项目注册自己的模块信息，在项目进行初始化时，会按照模块的依赖关系进行初始化。

用户可以在模块信息中添加当前模块的路由信息、服务、计划任务等，
这些功能在模块初始化时进行统一的注册初始化。

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

#### 插件

web 支持以插件的形式分发模块内容，只需要以 `-buildmode=plugin` 的形式编译每个模块，
之后将编译后的模块文件放至 `Config.Plugins` 配置项所指定的目录下即可。具体的可参考下 internal/plugintest 下的插件示例。

Go 并不是在所有的平台下都支持插件模式，支持列表可查看：<https://golang.org/pkg/plugin/>

字符集和文档类型
---

文档类型由 `Config.Marshalers` 和 `Config.Unmarshalers` 两个选项指定，分别对应编码和解码。
字符类型无需用户指定，<https://www.iana.org/assignments/character-sets/character-sets.xhtml>
中列出的字符集都能自动转换。

```go
conf := &web.Config {
    Marshalers: map[string]contentype.MarhsalFunc{
        "application/json": json.Marshal,
        "application/xml": xml.Marshal,
    },
    Unmarshalers: map[string]contentype.UnmarhsalFunc{
        "application/json": json.Unmarshal,
        "application/xml": xml.Unmarshal,
    },
    // 其它设置项
}

srv := web.New(conf)
srv.Serve()
```

客户端只要在请求时设置 Accept 报头就可返回相应类型的数据，而 Accept-Charset 报头可设置接收的字符集。
Content-Type 则可以有向服务器指定提交内容的文档类型和字符集。

错误处理
---

框架提供了一种输出错误信息内容的机制，用户只需要实现 Result 接口，即可自定义输出的错误信息格式。
具体实现可参考 context.defaultResult 的实现。

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
