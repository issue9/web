// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/sliceutil"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/config"
	"github.com/issue9/web/content"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Command 提供一种简单的命令行处理方式
//
// 由 Command 生成的命令行带以下三个参数：
//  - tag 运行的标签；
//  - v 显示版本号；
//  - fs 指定当前程序可读取的文件目录；
// 以上三个参数的参数名称，可在配置内容中修改。
//
//  cmd := &web.Command{
//      Name: "app",
//      Version: "1.0.0",
//      ServeTags: []string{"serve"},
//      InitServer: func(s *Server) error {...}
//  }
//
//  cmd.Exec()
type Command struct {
	Name string // 程序名称

	Version string // 程序版本

	// 在运行服务之前对 server 的额外操作，比如添加模块等。
	Init func(*Server) error

	// 当作服务运行的标签名
	//
	// 当标签名与此值相同时，在执行完 Server.InitModules 之后，还会执行 Server.Serve。
	//
	// 可以为空。
	ServeTags []string

	// 触发退出的信号
	//
	// 为空(nil 或是 []) 表示没有。
	Signals []os.Signal

	// 自定义命令行参数名
	CmdVersion string // 默认为 v
	CmdTag     string // 默认为 tag
	CmdFS      string // 默认为 fs

	// 命令行输出信息的通道
	//
	// 可以 os.Stdout 和 os.Stderr 选择，默认为 os.Stdout。
	Out io.Writer

	// 以下是初始 Server 对象的参数

	ResultBuilder content.BuildResultFunc // 默认为 nil，最终会被初始化 content.DefaultBuilder
	Locale        *serialization.Locale   // 默认情况下，能正常解析 xml、yaml 和 json
	LogsFilename  string                  // 默认为 logs.xml
	WebFilename   string                  // 默认为 web.yaml
}

// Exec 执行命令行操作
func (cmd *Command) Exec() {
	if err := cmd.sanitize(); err != nil {
		panic(err)
	}

	if err := cmd.exec(); err != nil {
		panic(err)
	}
}

func (cmd *Command) sanitize() *config.Error {
	if cmd.Name == "" {
		return &config.Error{Field: "Name", Message: "不能为空"}
	}
	if cmd.Version == "" {
		return &config.Error{Field: "Version", Message: "不能为空"}
	}

	if cmd.Init == nil {
		return &config.Error{Field: "InitServer", Message: "不能为空"}
	}

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	if cmd.Locale == nil {
		l := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))

		if err := l.Files().Add(json.Marshal, json.Unmarshal, ".json"); err != nil {
			return &config.Error{Field: "Locale", Message: err.Error()}
		}

		if err := l.Files().Add(xml.Marshal, xml.Unmarshal, ".xml"); err != nil {
			return &config.Error{Field: "Locale", Message: err.Error()}
		}

		if err := l.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"); err != nil {
			return &config.Error{Field: "Locale", Message: err.Error()}
		}

		cmd.Locale = l
	}

	if cmd.LogsFilename == "" {
		cmd.LogsFilename = "logs.xml"
	}
	if cmd.WebFilename == "" {
		cmd.WebFilename = "web.yaml"
	}

	if cmd.CmdFS == "" {
		cmd.CmdFS = "fs"
	}
	if cmd.CmdTag == "" {
		cmd.CmdTag = "tag"
	}
	if cmd.CmdVersion == "" {
		cmd.CmdVersion = "v"
	}

	return nil
}

func (cmd *Command) exec() error {
	v := flag.Bool(cmd.CmdVersion, false, "显示版本号")
	tag := flag.String(cmd.CmdTag, "", "执行的标签")
	f := flag.String(cmd.CmdFS, "./", "可读取的目录")
	flag.Parse()

	if *v {
		_, err := fmt.Fprintln(cmd.Out, cmd.Name, cmd.Version)
		return err
	}

	srv, err := LoadServer(cmd.Name, cmd.Version, cmd.ResultBuilder, cmd.Locale, os.DirFS(*f), cmd.LogsFilename, cmd.WebFilename)
	if err != nil {
		return err
	}

	if err := cmd.Init(srv); err != nil {
		return err
	}

	if err := srv.InitModules(*tag); err != nil {
		return err
	}

	// 如果不需要运行服务，则在这里退出。
	if sliceutil.Index(cmd.ServeTags, func(i int) bool { return cmd.ServeTags[i] == *tag }) < 0 {
		return nil
	}

	if len(cmd.Signals) > 0 {
		cmd.grace(srv, cmd.Signals...)
	}
	return srv.Serve()
}

func (cmd *Command) grace(s *server.Server, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := s.Close(3 * time.Second); err != nil {
			s.Logs().Error(err)
		}
		s.Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}
