web [![Build Status](https://travis-ci.org/issue9/web.svg?branch=master)](https://travis-ci.org/issue9/web)
======

**实验性的内容，勿用！**

如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

web 包是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能，当前也会有更多的限制。


#### 配置文件

通过 web.Init() 函数，可以在初始化时指定配置文件所在的目录，目前 web 包本身需要一个配置文件 `web.json`
以下是该文件的所有配置项：

| 名称                   | 类型   | 描述
|:-----------------------|:-------|:-----
| root                   | string | 项目的根路径，比如 `https://caixw.io/root/`
| server                 | object | 与 http 服务相关的设置
| server.https           | bool   | 是否启用 HTTPS
| server.httpState       | string | 当启用 HTTPS 时，针对 80 端口的处理方式，可以是 disable：不作任何处理；listen：与 https 作相同的算是；redirect 跳转到 https 相对应的端口。
| server.certFile        | string | 当启用 HTTPS 时的 cert 文件
| server.keyFile         | string | 当启用 HTTPS 时的 key 文件
| server.port            | string | 监听端口，以冒号(:) 开头
| server.headers         | object | 输出的报头，键名为报头名称，键值为对应的值
| server.static          | object | 静态内容，键名为 URL 地址，键值为对应的文件夹
| server.options         | bool   | 是否启用 OPTIONS 请求方法，默认为启用
| server.version         | string | 是否所有的接口只限定此版本，版本号在 accept 报头中指定，格式为 value=xx;version=xx
| server.hosts           | array  | 限定域名，可以是多个，若不指定，表示不限定
| server.readTimeout     | int    | 与 http.Server.ReadTimeout 相同，单位*纳秒*
| server.writeTimeout    | int    | 与 http.Server.WriteTimeout 相同，单位*纳秒*
| server.pprof           | string | pprof 的相关调试的地址，若为空，表示没有。
| content                | object | 与读取内容相关的各类操作
| content.contentType    | string | 默认的编码方式，目前 web 仅支持 json 和 xml 两种方式
| content.envelopeState  | string | envelope 的状态，可选值为： disable: 不启用；enable: 启用；must：强制启用，有着 envelope 的相关信息可查看之后的内容
| content.envelopeKey    | string | 当 envelopeState 的值为 enable 时，通过在地址中传递此值表示是否启用，默认值为 envelope
| content.envelopeStatus | string | 当 envelope 不为 disable 时，请求时返回的状态值，默认 200



#### 日志处理

日志处理，采用 [logs](https://github.com/issue9/logs) 包，一旦 web.Init() 或是 web.NewApp() 调用，logs 包即是处于可用状态。
logs 的配置文件与 `web.json` 一样放在同一目录下，可根据需求自行修改。

web.Context 提供了一套与 logs 相同接口的日志处理方法，相对于直接调用 logs，这些方法可以输出更多的调试信息，但其底层还是调用
logs 完成相应功能。


#### ContentType

目前支持 `json` 和 `xml` 两种编码方式，均以官方标准库的处理方式进行处理，服务端的渲染均支持这两种方式，
包括 envelope  和 Result 的输出，用户自定义内容需要自行解决编码问题。

默认情况下通过配置文件中的 server.contentType 值来获取当前的编码方式，用户也可以通过 NewContext()
中的第三个参数来自定义相应的编码。

**若项目中要支持 xml，需要注意，标准库中的 xml 包不支持 map 形式的数据转换成 xml**

#### envelope

envelope 模式主要是在部分客户端无法获取报头内容的情况下，将所有报头内容以报文的形式输出的兼容模式。
比如 JSON 格式会输出以下格式，response 为实际输出的内容：
```json
{
    "statue": 200,
    "headers": [
        {"ContentType": "application/json"},
        {"ContentLen": "1024"}
    ],
    "response": {
        "count": 24,
        "list": [
            {}
        ]
    }
}
```

在配置文件中可以指定 envelope 的状态：

1. disable 表示禁用；
1. must 表示强制启用；
1. enable 表示自动选择。

当值为 `enable` 时，只有通过传递查询参数，才能将返回内容转换成 envelope 模式，比如：
`https://caixw.io/users?envelope=true` 其中 envelope 可以通过 `envelopeKey` 配置项指定，
但值必须为 true 才表示启用。



#### 错误处理

框架中定义了一个统一的错误返回类型：Result，其输出格式是固定的，类似以下：
```json
{
    "code": 400001,
    "message": "error message",
    "detail": [
        {"field": "username", "message": "不能为空"},
        {"field": "password", "message": "不能为空"},
    ]
}
```

具体可参考代码文档中的有关 Result 的定义。


### 安装

```shell
go get github.com/issue9/web
```


### 文档

[![Go Walker](https://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/web)
[![GoDoc](https://godoc.org/github.com/issue9/web?status.svg)](https://godoc.org/github.com/issue9/web)


### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
