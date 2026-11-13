# web

[![Test](https://github.com/issue9/web/actions/workflows/test.yml/badge.svg)](https://github.com/issue9/web/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/issue9/web/graph/badge.svg?token=D5y3FOJk8A)](https://codecov.io/gh/issue9/web)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/web)](https://pkg.go.dev/github.com/issue9/web)
![Go version](https://img.shields.io/github/go-mod/go-version/issue9/web)
![License](https://img.shields.io/github/license/issue9/web)

web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能。
如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

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

## 字符集和文档类型

<https://www.iana.org/assignments/character-sets/character-sets.xhtml> 中列出的字符集都能自动转换。
支持的文档类型由 `web.Codec` 注册和配置。

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

## 错误处理

框架根据 [RFC7807](https://datatracker.ietf.org/doc/html/rfc7807) 提供了一种输出错误信息内容的机制。

在处理出错时，调用 Context.Problem 即可：

```go
func getAdmins(ctx *web.Context) web.Responser {
	return ctx.Problem(web.ProblemBadRequest).WithParam("param", "invalid format")
}
```

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

## 插件

- <https://github.com/issue9/webuse> 提供了中间件、插件、过滤器等常用的功能。

## 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
