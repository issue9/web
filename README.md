# web

[![Test](https://github.com/issue9/web/workflows/Test/badge.svg)](https://github.com/issue9/web/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/web)](https://goreportcard.com/report/github.com/issue9/web)
[![codecov](https://codecov.io/gh/issue9/web/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/web)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/web)](https://pkg.go.dev/github.com/issue9/web)
![Go version](https://img.shields.io/github/go-mod/go-version/issue9/web)
![License](https://img.shields.io/github/license/issue9/web)

web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能。
如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

```go
package main

import "github.com/issue9/web"

// main.go
func main() {
    srv := web.NewServer("web", "1.0.0", &web.Options{})
    router := srv.Routers().NewRouter()
    router.Get("/admins", getAdmins).
        Get("/groups", getGroups)

    srv.Serve()
}

func getAdmins(ctx* web.Context) web.Responser {
    return ctx.NotImplemented();
}

func getGroups(ctx* web.Context) web.Responser {
    return ctx.NotImplemented();
}
```

## 字符集和文档类型

文档类型由 `Server.Mimetypes` 指定。
字符类型无需用户指定，<https://www.iana.org/assignments/character-sets/character-sets.xhtml>
中列出的字符集都能自动转换。

```go
import "github.com/issue9/web"

srv := web.NewServer("app", "1.0.0", &web.Options{})

srv.Mimetypes().Add("application/json", json.Marshal, json.Unmarshal)
srv.Mimetypes().Add("application/xml", xml.Marshal, xml.Unmarshal)

srv.Serve()
```

客户端只要在请求时设置 Accept 报头就可返回相应类型的数据，而 Accept-Charset 报头可设置接收的字符集。
Content-Type 则可以有向服务器指定提交内容的文档类型和字符集。

## 错误处理

框架提供了一种输出错误信息内容的机制，用户只需要实现 Problem 接口，即可自定义输出的错误信息格式。
可参考内置的 [RFC7807](https://datatracker.ietf.org/doc/html/rfc7807) 实现：`internal/problems.rfc7807`。

## 中间件

- <https://github.com/issue9/middleware> 提供了部分中间件。
- <https://github.com/issue9/validator> 提供了部分验证方法。

## 工具

<https://github.com/issue9/web/releases> 提供了一个简易的辅助工具。可以帮助用户完成以下工作：

- 提取和更新本地化信息；
- 生成 openapi 文档。需要在注释中写一定的注解；

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
