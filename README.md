web [![Build Status](https://travis-ci.org/issue9/web.svg?branch=master)](https://travis-ci.org/issue9/web)
======

**实验性的内容，勿用！**

这是一个采用 Go 开发的 API 开发框架。
如果你只是需要一个简单的路由工具，那么你可以移步到 [mux](https://github.com/issue9/mux)。


#### 配置文件 web.json

TODO


#### 日志处理

TODO


#### content-type

目前支持 `json` 和 `xml` 两种编码方式，均以官方标准库的处理方式进行处理，
TODO

###### 注意事项

若项目需要支持 xml 返回，需要注意，xml 包不支持 map 形式的数据转换成 xml


#### envelope

TODO



#### 错误处理

框架中定义了一个错误返回类型：Result，
TODO


### 安装

```shell
go get github.com/issue9/web
```


### 文档

[![Go Walker](https://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/web)
[![GoDoc](https://godoc.org/github.com/issue9/web?status.svg)](https://godoc.org/github.com/issue9/web)


### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
