CONTRIBUTING
===


### 第三方包：

原则上不引用除 golang.org/x 和 github.com/issue9 之外的其它包。



### 测试：

包含了部分 `go generate` 内容，所以测试之前，需要调用以下命令生成 so 文件：
```shell
go generate ./...
```
或是执行根目录下的 `test.sh` 或 `test.ps1` 文件运行测试内容。
