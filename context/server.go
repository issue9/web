// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/compress"
	"github.com/issue9/middleware/v2/debugger"
	"github.com/issue9/middleware/v2/errorhandler"
	"github.com/issue9/middleware/v2/header"
	"github.com/issue9/middleware/v2/host"
	"github.com/issue9/mux/v2"
)

// Server 定义了构建 Context 对象的一些通用数据选项
type Server struct {
	// Interceptor 可以对生成的 Context 在使用前进行修改
	Interceptor func(*Context)

	// Location 指定服务器的时区信息
	//
	// 如果未指定，则会采用 time.Local 作为默认值。
	// 在构建 Context 对象时，该时区信息也会分配给 Context，
	// 如果每个 Context 对象需要不同的值，可以在 Interceptor 中进行修改。
	Location *time.Location

	// middleware
	middlewares   *middleware.Manager
	headers       *header.Header
	domains       *host.Host
	compress      *compress.Compress
	errorHandlers *errorhandler.ErrorHandler
	debugger      *debugger.Debugger

	// url
	root   string
	url    *url.URL
	router *mux.Prefix

	logs   *logs.Logs
	uptime time.Time

	// result
	resultBuilder BuildResultFunc
	messages      map[int]*resultMessage

	// mimetype
	marshals   []*marshaler
	unmarshals []*unmarshaler
}

// NewServer 返回 *Server 实例
func NewServer(logs *logs.Logs, builder BuildResultFunc, disableOptions, disableHead bool, root string) (*Server, error) {
	if builder == nil {
		builder = DefaultResultBuilder
	}

	// 保证不以 / 结尾
	if len(root) > 0 && root[len(root)-1] == '/' {
		root = root[:len(root)-1]
	}

	u, err := url.Parse(root)
	if err != nil {
		return nil, err
	}

	mux := mux.New(disableOptions, disableHead, false, nil, nil)
	router := mux.Prefix(u.Path)

	srv := &Server{
		Location: time.Local,

		middlewares: middleware.NewManager(router.Mux()),
		headers:     header.New(nil, nil),
		domains:     host.New(true),
		compress: compress.New(logs.ERROR(), map[string]compress.WriterFunc{
			"gzip":    compress.NewGzip,
			"deflate": compress.NewDeflate,
			"br":      compress.NewBrotli,
		}, "*"),
		errorHandlers: errorhandler.New(),
		debugger:      &debugger.Debugger{},

		root:   root,
		url:    u,
		router: router,

		logs:   logs,
		uptime: time.Now(),

		resultBuilder: builder,
		messages:      make(map[int]*resultMessage, 20),

		marshals:   make([]*marshaler, 0, 10),
		unmarshals: make([]*unmarshaler, 0, 10),
	}

	srv.buildMiddlewares()

	return srv, nil
}

// Logs 返回关联的 logs.Logs 实例
func (srv *Server) Logs() *logs.Logs {
	return srv.logs
}

// AddStatic 添加静态路由
//
// 键名为 URL 的路径部分，相对于项目根路径，键值为文件地址。
//
// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
func (srv *Server) AddStatic(path, dir string) {
	h := http.StripPrefix(path, http.FileServer(http.Dir(dir)))
	srv.Router().Get(path+"{path}", h)
}

// AllowedDomain 添加允许访问的域名
//
// 若指定了此值，则只有此列表中指定的域名可以访问当前网页。
// 诸如 IP 和其它域名的指向将不再启作用。
//
// 可以指定泛域名，比如 *.example.com
func (srv *Server) AllowedDomain(domain ...string) {
	// NOTE: 如果要检测域名的合法性，请注意 *.example.com 的泛域名格式。
	srv.domains.Add(domain...)
}

// SetHeader 附加的报头信息
//
// 一些诸如跨域等报头信息，可以在此作设置。
//
// NOTE: 报头信息可能在其它处理器被修改。
//
// name 表示报头名称；value 表示报头的值，如果为空，表示删除该报头。
func (srv *Server) SetHeader(name, value string) {
	if value == "" {
		srv.headers.Delete(name)
	} else {
		srv.headers.Set(name, value)
	}
}

// SetErrorHandle 设置指定状态码页面的处理函数
//
// 如果状态码已经存在处理函数，则修改，否则就添加。
// 仅对状态码 >= 400 的有效果。
func (srv *Server) SetErrorHandle(h errorhandler.HandleFunc, status ...int) {
	srv.errorHandlers.Set(h, status...)
}

// AddMiddlewares 设置全局的中间件
func (srv *Server) AddMiddlewares(m middleware.Middleware) {
	srv.middlewares.After(m)
}

// SetDebugger 设置调试地址
func (srv *Server) SetDebugger(pprof, vars string) {
	srv.debugger.Pprof = pprof
	srv.debugger.Vars = vars
}

// Handler 将当前服务转换为 http.Handler 接口对象
func (srv *Server) Handler() http.Handler {
	return srv.middlewares
}

// 通过配置文件加载相关的中间件
//
// 始终保持这些中间件在最后初始化。用户添加的中间件由 app.modules.After 添加。
func (srv *Server) buildMiddlewares() {
	srv.middlewares.Before(srv.headers.Middleware)

	srv.middlewares.Before(srv.domains.Middleware)

	srv.middlewares.Before(srv.errorHandlers.Middleware)

	// srv.errorhandlers.New 可能会输出大段内容。所以放在其之后。
	srv.middlewares.Before(srv.compress.Middleware)

	// recovery
	rf := srv.errorHandlers.Recovery(srv.logs.ERROR())
	srv.middlewares.Before(rf.Middleware)

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	srv.middlewares.Before(srv.debugger.Middleware)
}

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time {
	return srv.uptime
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time {
	return time.Now().In(srv.Location)
}

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location)
}

// Server 获取关联的 context.Server 实例
func (ctx *Context) Server() *Server {
	return ctx.server
}

// Path 生成路径部分的地址
func (srv *Server) Path(p string) string {
	p = path.Join(srv.url.Path, p)
	if p != "" && p[0] != '/' {
		p = "/" + p
	}

	return p
}

// URL 构建一条基于 Root 的完整 URL
func (srv *Server) URL(p string) string {
	switch {
	case len(p) == 0:
		return srv.root
	case p[0] == '/':
		// 由 NewServer 保证 root 不能 / 结尾
		return srv.root + p
	default:
		return srv.root + "/" + p
	}
}
