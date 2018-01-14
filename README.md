web [![Build Status](https://travis-ci.org/issue9/web.svg?branch=master)](https://travis-ci.org/issue9/web)
[![Go version](https://img.shields.io/badge/Go-1.8-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/web)](https://goreportcard.com/report/github.com/issue9/web)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
======

**实验性的内容，勿用！**

web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能，当前也会有更多的限制。

如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

```go
// main.go
func main() {
    opt := &web.Options {
        ConfigDir: "./appconfig",
        Charset: map[string]encoding.Encoding {
            "gb18030": simplifiedchinese.GB18030,
            "gbk": simplifiedchinese.GBK,
        },
        Marshals: map[string]context.Marshaler {
            "application/json": json.Marshal,
            "application/xml": xml.Marshal,
        },
        Unmarshals: map[string]context.Unmarshaler {
            "application/json": json.Unmarshal,
            "application/xml": xml.Unmarshal,
        }
    }

    web.Init(opt)

    // 注册模块信息
    web.AddModule(m1.Module)
    web.AddModule(m2.Module)

    logs.Fatal(web.Run(nil))
}

// m1/module.go
var Module = web.NewModule("m1", "模块描述信息").
    GetFunc("/admins", getAdmins).
    GetFunc("/groups", getGroups)

// m2/module.go
var Module = web.NewModule("m2", "模块描述信息", "m1").
    GetFunc("/admins", getAdmins).
    GetFunc("/groups", getGroups)
```


#### 配置文件

通过 web.Init() 函数，可以在初始化时指定配置文件所在的目录，目前 web 包本身需要一个配置文件 `web.yaml`
以下是该文件的所有配置项：

| 名称            | 类型   | 描述
|:----------------|:-------|:-----
| debug           | bool   | 是否启用调试模式
| domain          | string | 项目的域名，若存在 allowedDomains 同时会加入到 allowedDomains 字段中
| root            | string | 项目的根路径，比如 `/blog`
| outputEncoding  | string | 默认的编码方式
| outputCharset   | string | 字符集
| strict          | bool   | 启用此值，会检测用户的 Accept 报头是否符合当前的编码。
| https           | bool   | 是否启用 HTTPS
| httpState       | string | 当启用 HTTPS 时，针对 80 端口的处理方式，可以是 disable：不作任何处理；listen：与 https 作相同的算是；redirect 跳转到 https 相对应的端口。
| certFile        | string | 当启用 HTTPS 时的 cert 文件
| keyFile         | string | 当启用 HTTPS 时的 key 文件
| port            | string | 监听端口，以冒号(:) 开头
| headers         | object | 输出的报头，键名为报头名称，键值为对应的值
| static          | object | 静态内容，键名为 URL 地址，键值为对应的文件夹
| options         | bool   | 是否启用 OPTIONS 请求方法，默认为启用
| version         | string | 是否所有的接口只限定此版本，版本号在 accept 报头中指定，格式为 value=xx;version=xx
| allowedDomains  | array  | 限定访问域名，可以是多个，若不指定，表示不限定
| plugins         | array  | 指定一组 so 文件，可以当作插件的形式进行加载
| readTimeout     | string | 与 http.Server.ReadTimeout 相同
| writeTimeout    | string | 与 http.Server.WriteTimeout 相同



#### 日志处理

日志处理，采用 [logs](https://github.com/issue9/logs) 包，一旦 web.Init() 或是 web.NewApp() 调用，logs 包即是处于可用状态。
logs 的配置文件与 `web.json` 一样放在同一目录下，可根据需求自行修改。

web.Context 提供了一套与 logs 相同接口的日志处理方法，相对于直接调用 logs，这些方法可以输出更多的调试信息，但其底层还是调用
logs 完成相应功能。


#### ContentType

输出的编码方式与字符集由用户在配置文件中指定，而输入的编码方式与字符集，
由客户端在请求时，通过 `Content-Type` 报头指定。当然系统具体可以支持什么编码和字符集，
由用户在开始前通过 `AddMarshal()`、`AddUnmarshal()` 和 `AddCharset()`
来指定一个列表，在此列表内的编码和字符集，均可用。



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


### 用到的第三方包

- yaml gopkg.in/yaml.v2 配置文件使用 yaml 格式，比 JSON 拥有更好的阅读体验；
- text golang.org/x/text 提供了非 UTF-8 字符集的转码方式。


### 文档

[![Go Walker](https://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/web)
[![GoDoc](https://godoc.org/github.com/issue9/web?status.svg)](https://godoc.org/github.com/issue9/web)


### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
