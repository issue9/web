CONTRIBUTING
===


### 第三方包：

- yaml gopkg.in/yaml.v2 配置文件使用 yaml 格式，比 JSON 拥有更好的阅读体验；
- text golang.org/x/text 提供了非 UTF-8 字符集的转码方式。

原则上不会引入除 golang.org/x 和 github.com/issue9 之外的其它包。




### 测试：

包含了部分 `go generate` 内容，所以测试之前，需要调用以下命令生成 so 文件：
```shell
go generate ./...
```
或是执行根目录下的 `test.sh` 或 `test.ps1` 文件运行测试内容。
