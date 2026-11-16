# web

[![Test](https://github.com/issue9/web/actions/workflows/test.yml/badge.svg)](https://github.com/issue9/web/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/issue9/web/graph/badge.svg?token=D5y3FOJk8A)](https://codecov.io/gh/issue9/web)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/web)](https://pkg.go.dev/github.com/issue9/web)
![Go version](https://img.shields.io/github/go-mod/go-version/issue9/web)
![License](https://img.shields.io/github/license/issue9/web)

web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能。
如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

## 安装

项目要求 Go 1.27 或更高版本，可通过 Go Modules 添加依赖：

```bash
go get github.com/issue9/web
```

```go
package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func main() {
	srv, err := server.NewHTTP("web", "1.0.0", &server.Options{
		Startup: web.NewHTTPServerStartup(&http.Server{Addr: ":8080"}),
	})
	if err != nil {
		log.Fatal(err)
	}

	router := srv.Routers().New("default", nil)
	router.Get("/admins", getAdmins).
		Get("/groups", getGroups)

	if err := srv.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func getAdmins(ctx *web.Context) web.Responser {
	return ctx.NotImplemented()
}

func getGroups(ctx *web.Context) web.Responser {
	return ctx.NotImplemented()
}
```

## 编码

编码相关的功能由 `Codec` 结构体管理，总体上分为以下几个类型：

### 字符集和文档类型

并没有像其它框架一样为不同的媒体类型输出提供各自不同的方法，而返回统一的对象，
再由客户端的 Accept 等报头协商返回给客户端需要的类型。比如要向用户返回一个 400 的错误，
只需要以下代码：

```go
func handler(ctx *web.Context) web.Responser {
    return ctx.Problem(problems.BadRequest).WithParam("username", "不能为空")
}
```

以上代码会向客户端返回 400 的错误，且带一个 username 不能为空的错误信息。
但是具体是渲染成 JSON 还是 XML 亦或是 HTML，无须 handler 函数操心，
可以根据客户端的 Accept 报头提供客户端最需要的格式，
比如客户端的 Accept 为 `application/json;q=1` 那么会返回 JSON 格式，
如果 Accept 为 `application/json;q=0.5,application/xml;q=1` 那么会返回 XML 格式。
服务端只需要在初始化服务时配置对应的 `web.Codec` 对象就可以了。

字符集也是如此，甚至不需要服务端做配置，本身就支持在 <https://www.iana.org/assignments/character-sets/character-sets.xhtml> 中列出的所有字符集。

```go
import (
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/codec"
)

func newServer() (*web.Server, error) {
	c := web.NewCodec().
		AddMediaType(codec.NewJSON(), true, true).
		AddMediaType(codec.NewXML(), true, true)

	return server.NewHTTP("app", "1.0.0", &server.Options{Codec: c})
}
```

客户端只要在请求时设置 Accept 报头就可返回相应类型的数据，而 Accept-Charset 报头可设置接收的字符集。
Content-Type 则可以有向服务器指定提交内容的文档类型和字符集。

### 压缩

通过 `web.Codec.AddCompressor` 可以添加需要支持压缩算法。`codec` 提供大部分目前浏览器支持的压缩算法。
使用方法也同媒体类型，通过客户端的 `Accept-Encoding` 自动匹配压缩算法。

## 错误处理

框架根据 [RFC7807](https://datatracker.ietf.org/doc/html/rfc7807) 和 [RFC9457](https://datatracker.ietf.org/doc/html/rfc9457) 提供了一种输出错误信息内容的机制。

在处理出错时，调用 Context.Problem 即可：

```go
import (
	"github.com/issue9/web"
	"github.com/issue9/web/problems"
)

func getAdmins(ctx *web.Context) web.Responser {
	return ctx.Problem(problems.BadRequest).WithParam("param", "invalid format")
}
```

### 过滤器

实现 `web.ContextFilter` 接口即可在请求数据解码后验证或修正数据。下面的例子会先去掉用户名首尾空格，再检查用户名非空和年龄范围：

```go
import (
	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/filter/sanitizers"
	"github.com/issue9/web/filter/validators"
	"github.com/issue9/web/problems"
)

type CreateUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (u *CreateUser) FilterContext(f *filter.Filter, _ *web.Context) {
	f.Sanitize(&u.Name, sanitizers.Trim)
	f.Valid("name", u.Name, validators.Not(validators.Empty), web.Phrase("不能为空"))
	f.Valid("age", u.Age, validators.BetweenEqual(18, 120), web.Phrase("年龄必须在 18 到 120 之间"))
}

func createUser(ctx *web.Context) web.Responser {
	var input CreateUser
	if resp := ctx.Read(false, &input, problems.BadRequest); resp != nil {
		return resp
	}

	// input 已通过验证，可以继续处理。
	return web.NoContent()
}
```

`ctx.Read` 在请求体解码成功后调用 `FilterContext`；查询参数对象也会在 `ctx.QueryObject` 解码后执行该方法。`exitAtError` 为 `false` 时会收集所有验证错误，为 `true` 时遇到首个错误即停止。
`filter/validators` 和 `filter/sanitizers` 提供常用的验证器与修正函数，也可以在 `FilterContext` 中组合自定义逻辑。

## openapi

可直接在添加 API 的中间件上指定文档内容：

```go
import (
	"github.com/issue9/web"
	"github.com/issue9/web/openapi"
)

func registerRoutes(srv *web.Server) {
	router := srv.Routers().New("default", nil)
	doc := openapi.New(srv, web.Phrase("title"))

	router.Get("/admins", getAdmins, doc.API(func(o *openapi.Operation) {
		o.Desc(web.Phrase("desc of API"))
	}))
	router.Get("/openapi", doc.Handler())
}
```

## 初始化

提供了多种初始化项目的方式，最简单的方式是直接使用 `web.NewServer` 创建 `Server` 对象，但是这个需要提供多个参数，
且要自己保证数据的正确性，一般不推荐。

在 `server` 下提供了几种常用的初始化方式：

- `server` 通过一个可选的 `Options` 对象初始化 `Server`；
- `server/config` 提供了从配置文件加载 `server.Options` 的功能，配置项详见 [配置文档](server/config/README.md)；
- `server/app` 则提供了直接创建命令行的工具，可以实现对 `Server` 的重启以及注册成系统服务等功能；

## 插件

- <https://github.com/issue9/webuse> 提供了中间件、插件、过滤器等常用的功能。

## 参与项目

欢迎通过 Issue 和 Pull Request 参与项目。提交前请阅读[贡献指南](CONTRIBUTING.md)，并遵守[行为准则](CODE_OF_CONDUCT.md)。
尚未修复的安全问题请按照[安全政策](SECURITY.md)私下报告，不要在公开 Issue 中披露。

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
