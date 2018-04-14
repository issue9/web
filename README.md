web [![Build Status](https://travis-ci.org/issue9/web.svg?branch=master)](https://travis-ci.org/issue9/web)
[![Go version](https://img.shields.io/badge/Go-1.10-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/web)](https://goreportcard.com/report/github.com/issue9/web)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/web/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/web)
======


web 是一个比较完整的 API 开发框架，相对于简单的路由，提供了更多的便利功能。

如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。

```go
// main.go
func main() {
    // 可以自动处理 content-type 的值为 charset=gb18083 和 gbk 的请求，会自动转码
    encoding.AddCharset(map[string]encoding.Encoding {
        "gb18030": simplifiedchinese.GB18030,
        "gbk": simplifiedchinese.GBK,
    })

    encoding.AddMarshals(map[string]context.Marshaler {
        "application/json": json.Marshal,
        "application/xml": xml.Marshal,
    })

    encoding.AddUnmarshals(map[string]context.Unmarshaler {
        "application/json": json.Unmarshal,
        "application/xml": xml.Unmarshal,
    })

    result.NewMessages(map[int]string{...})

    web.Init("./appconfig", nil)

    // 注册模块信息
    m1.Init()
    m2.Init()

    logs.Fatal(web.Run())
}

// modules/m1/module.go
func Init() {
    web.NewModule("m1", "模块描述信息").
        GetFunc("/admins", getAdmins).
        GetFunc("/groups", getGroups)
}

// modules/m2/module.go
func Init() {
    web.NewModule("m2", "模块描述信息", "m1").
        GetFunc("/admins", getAdmins).
        GetFunc("/groups", getGroups)
}
```


#### 项目结构

这只是推荐的目录结构，但不是必须按照此来。
```
|----- appconfig 配置文存放路径
|         |
|         +----- web.yaml 框架本身的配置文件
|         |
|         +----- logs.xml 日志配置文件
|
+----- common 一些公用的包
|
+----- modules 各个模块的代码
|         |
|         +----- module1
|         |
|         +----- module2
|
+----- main.go
```


#### 配置文件

通过 web.Init() 函数，可以在初始化时指定配置文件所在的目录，目前 web 包本身需要一个配置文件 `web.yaml`
以下是该文件的所有配置项：

| 名称            | 类型   | 描述
|:----------------|:-------|:-----
| debug           | bool   | 是否启用调试模式
| domain          | string | 项目的域名，若存在 allowedDomains 同时会加入到 allowedDomains 字段中
| root            | string | 项目的根路径，比如 `/blog`
| outputMimeType  | string | 默认输出媒体类型
| outputCharset   | string | 字符集
| strict          | bool   | 启用此值，会检测用户的 Accept 报头是否符合当前的编码。
| https           | bool   | 是否启用 HTTPS
| certFile        | string | 当启用 HTTPS 时的 cert 文件
| keyFile         | string | 当启用 HTTPS 时的 key 文件
| port            | string | 监听端口，以冒号(:) 开头
| headers         | object | 输出的报头，键名为报头名称，键值为对应的值
| static          | object | 静态内容，键名为 URL 地址，键值为对应的文件夹
| disableOptions  | bool   | 是否禁用 OPTIONS 请求方法
| allowedDomains  | array  | 限定访问域名，可以是多个，若不指定，表示不限定
| readTimeout     | string | 与 http.Server.ReadTimeout 相同
| writeTimeout    | string | 与 http.Server.WriteTimeout 相同
| shutdownTimeout | string | 关闭服务的等待时间

*详细的介绍可以参考 /internal/config/config.go 文件中的描述*



#### 日志处理

日志处理，采用 [logs](https://github.com/issue9/logs) 包，一旦 web.Init() 调用，logs 包即是处于可用状态。
logs 的配置文件与 `web.json` 一样放在同一目录下，可根据需求自行修改。

web.Context 提供了一套与 logs 相同接口的日志处理方法，相对于直接调用 logs，这些方法可以输出更多的调试信息，但其底层还是调用
logs 完成相应功能。



#### 字符集和媒体类型

输出的媒体类型与字符集由用户在配置文件中指定，而输入的媒体类型与字符集，
由客户端在请求时，通过 `Content-Type` 报头指定。
当然如果需要框架支持用户提交的类型，需要在框架初始化时，添加相关的编友支持：
由用户在开始前通过 `AddMarshal()`、`AddUnmarshal()` 和 `AddCharset()`
来指定一个列表，在此列表内的编码和字符集，均可用。



#### 错误处理

框架中提供了一个统一的错误返回类型：Result，其输出格式是固定的，类似以下：
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
