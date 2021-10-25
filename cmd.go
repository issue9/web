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
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/config"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Command 提供一种简单的命令行处理方式
//
// 由 Command 生成的命令行带以下三个参数：
//  - action 运行的标签；
//  - v 显示版本号；
//  - fs 指定当前程序可读取的文件目录；
// 以上三个参数的参数名称，可在配置内容中修改。
//
//  cmd := &web.Command{
//      Name: "app",
//      Version: "1.0.0",
//      ServeTags: []string{"serve"},
//      Init: func(s *Server) error {...},
//  }
//
//  cmd.Exec()
type Command struct {
	Name string // 程序名称

	Version string // 程序版本

	// 当作服务运行的标签名
	//
	// 当标签名在此列表时，Server.Serve 的第一个参数为 true。
	ServeTags []string

	// 触发退出的信号
	//
	// 为空(nil 或是 []) 表示没有。
	Signals []os.Signal

	// 错误代码以及对应的信息
	ResultMessages map[int]LocaleStringer

	// 在初始化 Server 之前对 Options 的二次处理
	//
	// 可以为空。
	Options OptionsFunc

	// 在运行服务之前对 Server 的额外操作
	//
	// 比如添加模块等。可以为空。
	Init func(*Server) error

	// 自定义命令行参数名
	CmdVersion string // 默认为 v
	CmdAction  string // 默认为 action
	CmdFS      string // 默认为 fs

	// 命令行输出信息的通道
	//
	// 可以 os.Stdout 和 os.Stderr 选择，默认为 os.Stdout。
	Out io.Writer

	// 以下是初始 Server 对象的参数

	Files          *serialization.Files // 为空初始化为能解析 .xml、.yaml、.yml 和 .json 文件的默认对象。
	ConfigFilename string               // 默认为 web.xml
}

// Exec 执行命令行操作
func (cmd *Command) Exec() error {
	if err := cmd.sanitize(); err != nil {
		return err
	}
	return cmd.exec()
}

func (cmd *Command) sanitize() *config.Error {
	if cmd.Name == "" {
		return &config.Error{Field: "Name", Message: "不能为空"}
	}
	if cmd.Version == "" {
		return &config.Error{Field: "Version", Message: "不能为空"}
	}

	if cmd.Init == nil {
		return &config.Error{Field: "Init", Message: "不能为空"}
	}

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	if cmd.Files == nil {
		f := serialization.NewFiles(5)

		if err := f.Add(json.Marshal, json.Unmarshal, ".json"); err != nil {
			return &config.Error{Field: "Files", Message: err}
		}

		if err := f.Add(xml.Marshal, xml.Unmarshal, ".xml"); err != nil {
			return &config.Error{Field: "Files", Message: err}
		}

		if err := f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"); err != nil {
			return &config.Error{Field: "Files", Message: err}
		}

		cmd.Files = f
	}

	if cmd.ConfigFilename == "" {
		cmd.ConfigFilename = "web.xml"
	}

	if cmd.CmdFS == "" {
		cmd.CmdFS = "fs"
	}
	if cmd.CmdAction == "" {
		cmd.CmdAction = "action"
	}
	if cmd.CmdVersion == "" {
		cmd.CmdVersion = "v"
	}

	return nil
}

func (cmd *Command) exec() error {
	v := flag.Bool(cmd.CmdVersion, false, "显示版本号")
	action := flag.String(cmd.CmdAction, "", "执行的标签")
	f := flag.String(cmd.CmdFS, "./", "可读取的目录")
	flag.Parse()

	if *v {
		_, err := fmt.Fprintln(cmd.Out, cmd.Name, cmd.Version)
		return err
	}

	if *action == "" {
		return &config.Error{Field: "action", Message: "不能为空"}
	}

	srv, err := LoadServer(cmd.Name, cmd.Version, cmd.Files, os.DirFS(*f), cmd.ConfigFilename, cmd.Options)
	if err != nil {
		return err
	}

	srv.AddResults(cmd.ResultMessages)

	if err := cmd.Init(srv); err != nil {
		return err
	}

	if len(cmd.Signals) > 0 {
		cmd.grace(srv, cmd.Signals...)
	}

	serve := sliceutil.Index(cmd.ServeTags, func(i int) bool { return cmd.ServeTags[i] == *action }) >= 0
	return srv.Serve(serve, *action)
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
