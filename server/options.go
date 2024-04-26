// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v8/group"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/locales"
	xj "github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

const DefaultConfigDir = "@.config" // 默认的配置目录地址

const (
	typeHTTP int = iota
	typeGateway
	typeService
)

type (
	// Options 初始化 [web.Server] 的参数
	//
	// NOTE: 这些参数都有默认值，且无法在 [web.Server] 初始化之后进行更改。
	Options struct {
		// 项目的配置文件管理
		//
		// 如果为空，则采用 [DefaultConfigDir] 作为配置文件的目录，
		// 同时加载 YAML、XML 和 JSON 三种文件类型的序列化方法。
		Config *config.Config

		// 服务器的时区
		//
		// 默认值为 [time.Local]
		Location *time.Location

		// 缓存系统
		//
		// 如果为空，采用 [github.com/issue9/cache/caches/memory/New] 作为默认值。
		Cache cache.Driver

		// 日志系统
		//
		// 如果此值为空，表示不会输出任何信息。
		//
		// 会调用 [logs.Logs.SetLocale] 设置为 [Language] 的值。
		Logs *logs.Logs

		// http.Server 实例的值
		HTTPServer *http.Server

		// 生成唯一字符串的方法
		//
		// 供 [Server.UniqueID] 使用。
		//
		// 如果为空，将采用 [unique.NewString] 作为生成方法。
		//
		// NOTE: 该值的修改，可能造成项目中的唯一 ID 不再唯一。
		IDGenerator func() string

		// 路由选项
		RoutersOptions []web.RouterOption

		// 指定获取 x-request-id 内容的报头名
		//
		// 如果为空，则采用 [header.XRequestID] 作为默认值
		RequestIDKey string

		// 编码方式
		//
		// 如果为空，则仅支持 JSON 编码，不支持压缩方式。
		Codec *web.Codec

		// 默认的语言标签
		//
		// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
		// 同时也用来初始化 [Server.Locale.Printer]。
		//
		// 框架中的日志输出时，如果该信息实现了 [web.LocaleStringer] 接口，
		// 将会转换成此设置项的语言。
		//
		// 如果为空，则会尝试读取当前系统的本地化信息。
		Language language.Tag

		locale *locale.Locale

		// 所有 [web.Problem.Type] 字段的前缀
		//
		// 如果该值为 [web.ProblemAboutBlank]，将不输出 ID 值；其它值则作为前缀添加。
		// 空值是合法的值，表示不需要添加前缀。
		ProblemTypePrefix string

		// OnRender 可实现对渲染结果的调整
		//
		// 默认为空。
		//
		// NOTE: 该值的修改，可能造成所有接口返回数据结构的变化。
		OnRender func(status int, body any) (int, any)

		// 指定对 [web.Server] 进行初始化的插件
		//
		// 这些插件会在 [web.Server.Serve] 运行之前被调用。
		Plugins []web.Plugin

		// 以下微服务相关的设置

		// 作为微服务时的注册中心实例
		//
		// NOTE: 仅在 [NewService] 和 [NewGateway] 中才会有效果。
		Registry registry.Registry

		// 作为微服务终端时的地址
		//
		// NOTE: 仅在 [NewService] 中才会有效果。
		Peer selector.Peer

		// 作为微服务网关时的 URL 映射关系
		//
		// NOTE: 仅在 [NewGateway] 中才会有效果。
		Mapper map[string]web.RouterMatcher
	}
)

func sanitizeOptions(o *Options, t int) (*Options, *web.FieldError) {
	if o == nil {
		o = &Options{}
	}

	if o.Config == nil {
		s := make(config.Serializer, 4)
		s.Add(json.Marshal, json.Unmarshal, ".json").
			Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml").
			Add(xml.Marshal, xml.Unmarshal, ".xml")

		c, err := config.BuildDir(s, DefaultConfigDir)
		if err != nil {
			return nil, web.NewFieldError("Config", err)
		}
		o.Config = c
	}

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.IDGenerator == nil {
		u := unique.NewString(1000)
		o.IDGenerator = u.String
		o.Plugins = append(o.Plugins, web.PluginFunc(func(s web.Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		}))
	}

	if o.Cache == nil {
		o.Cache = memory.New()
	}

	if o.Language == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			return nil, web.NewFieldError("Language", err)
		}
		o.Language = tag
	}
	o.locale = locale.New(o.Language, o.Config)

	if o.Logs == nil {
		o.Logs = logs.New(logs.NewNopHandler())
	}
	o.Logs.SetLocale(o.locale.Printer())

	if o.RequestIDKey == "" {
		o.RequestIDKey = header.XRequestID
	}

	if o.Codec == nil {
		o.Codec = web.NewCodec().AddMimetype(xj.Mimetype, xj.Marshal, xj.Unmarshal, xj.ProblemMimetype)
	}

	switch t {
	case typeHTTP: // 不需要处理任何数据
		return o, nil
	case typeGateway:
		return o, filter.ToFieldError(
			filter.New("Mapper", &o.Mapper, filter.V(func(v map[string]group.Matcher) bool { return v != nil }, locales.CanNotBeEmpty)),
			filter.New("Mapper", &o.Mapper, filter.MV[map[string]group.Matcher](func(v group.Matcher) bool { return v != nil }, locales.CanNotBeEmpty)),
			filter.New("Registry", &o.Registry, filter.V(func(v registry.Registry) bool { return v != nil }, locales.CanNotBeEmpty)),
		)
	case typeService:
		return o, filter.ToFieldError(
			filter.New("Peer", &o.Peer, filter.V(func(v selector.Peer) bool { return v != nil }, locales.CanNotBeEmpty)),
			filter.New("Registry", &o.Registry, filter.V(func(v registry.Registry) bool { return v != nil }, locales.CanNotBeEmpty)),
		)
	default:
		panic("参数 t 取值错误")
	}
}

func (o *Options) internalServer(name, version string, s web.Server) *web.InternalServer {
	return web.InternalNewServer(s, name, version,
		o.Location, o.Logs, o.IDGenerator, o.locale,
		o.Cache, o.Codec, o.RequestIDKey, o.ProblemTypePrefix,
		o.OnRender, o.RoutersOptions...)
}

// Render200 统一 API 的返回格式
//
// 适用 [Options.OnRender]。
//
// 返回值中，状态码统一为 [http.StatusOK]。返回对象统一为 [RenderResponse]。
func Render200(status int, body any) (int, any) {
	return http.StatusOK, &RenderResponse{OK: !web.IsProblem(status), Status: status, Body: body}
}

// RenderResponse API 统一的返回格式
type RenderResponse struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"body" cbor:"-"`
	OK      bool     `json:"ok" yaml:"ok" xml:"ok,attr" cbor:"ok"`                 // 是否是错误代码
	Status  int      `json:"status" yaml:"status" xml:"status,attr" cbor:"status"` // 原始的状态码
	Body    any      `json:"body" yaml:"body" xml:"body" cbor:"body"`
}

func (r *RenderResponse) MarshalHTML() (string, any) { return "render-response", r }
