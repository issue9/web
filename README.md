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
    conf := web.Classic("./appconfig/web.yaml")

    // 注册模块信息
    m1.Init()
    m2.Init()

    web.Serve()
}

// modules/m1/module.go
func Init(s *web.MODServer) {
    s.NewModule("m1", "模块描述信息").
        Get("/admins", getAdmins).
        Get("/groups", getGroups)
}

// modules/m2/module.go
func Init(s *web.MODServer) {
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
用于向框架注册当前模块的一些主要信息。通过 `web.MODServer` 注册模块：

```go
package m1

import "github.com/issue9/web"

func Init(s *web.MODServer) {
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

配置文件
---

通过 web.Classic() 函数，可以在初始化时指定配置文件，文件格式可以是 XML、JSON 和
YAML。用户也可以自行添加新的格式支持。

#### web.yaml

以下是该文件的所有配置项：

| 名称              | 类型   | 描述
|:------------------|:-------|:-----
| debug             | bool   | 是否启用调试模式
| root              | string | 项目的根路径，比如 `/blog`
| plugins           | string | 指定需要加载的插件，可以使用 glob 模式，仅支持部分系统，具体可见 https://golang.org/pkg/plugin/
| headers           | object | 输出的报头，键名为报头名称，键值为对应的值
| static            | object | 静态内容，键名为 URL 地址，键值为对应的文件夹
| disableOptions    | bool   | 是否禁用 OPTIONS 请求方法
| disableHead       | bool   | 是否禁用自动生成 HEAD 请求方法
| allowedDomains    | array  | 限定访问域名，可以是多个，若不指定，表示不限定
| readTimeout       | string | 与 http.Server.ReadTimeout 相同
| writeTimeout      | string | 与 http.Server.WriteTimeout 相同
| idleTimeout       | string | 与 http.Server.IdleTimeout 相同
| maxHeaderBytes    | int    | 与 http.Server.MaxHeaderBytes 相同
| readHeaderTimeout | string | 与 http.Server.ReadHeaderTimeout 相同
| shutdownTimeout   | string | 当请求关闭时，可用于处理剩余请求的时间。
| timezone          | string | 时区信息，名称为 IAAN 注册的名称，为空则为 Local
| certificates      | object | 多域名的证书信息

*详细的介绍可以参考 /internal/webconfig/webconfig.go 文件中的描述*

在 debug 模式下，会添加两个调试用的地址：`/debug/pprof/` 和 `/debug/vars`

#### logs.xml

`logs.xml` 采用 [logs](https://github.com/issue9/logs) 包的功能，具体的配置可参考其文档。

#### 字符集

字符集用户无须任何操作，会自动根据 `Content-Type` 中的 charset 属性自动解析其字符集，
输出时，也会根据 `Accept-Charset` 报头内容，作自动转换之后再输出。以下字符集都被支持：
<https://www.iana.org/assignments/character-sets/character-sets.xhtml>

#### 媒体类型

默认情况下，框架不会处理任何的 mimetype 类型的数据。需要用户通过
`web.CTXServer().AddMarshals()` 和 `web.CTXServer().AddUnmarshals()` 添加相关的处理函数。
添加方式如下：

```go
web := Classic("path")

web.CTXServer().AddMarshals(map[string]mimetype.MarshalFunc{
    "application/json": json.Marshal,
})
web.CTXServer().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
    "application/json": json.Unmarshal,
})
```

#### 错误处理

框架提供了一种输出错误信息内容的机制，用户只需要实现 Result 接口，
即可自定义输出的错误信息格式。

具体可参考代 result 中的相关代码。

安装
---

```shell
go get github.com/issue9/web
```

版权
---

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
