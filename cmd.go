// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Command 提供一种简单的命令行生成方式
//
// 由 Command 生成的命令行带以下几个参数：
//  - v 显示版本号；
//  - h 显示帮助信息；
//  - f 指定当前程序可读取的文件目录；
//  - i 执行安装脚本，如果参数不为空，则执行指定的标签；
//  - s 以服务运行，与 i 不能同时出现；
//
//  cmd := &web.Command{
//      Name: "app",
//      Version: "1.0.0",
//      Init: func(s *Server) error {...},
//  }
//
//  cmd.Exec()
//
// NOTE: Command 中的大部分内容在 LoadServer 之前就运行，所以无法适用 Server.Locale
// 的本地化操作，若有需要对命令行作本地化操作，需要自行实现。
type Command struct {
	Name string // 程序名称

	Version string // 程序版本

	// 在运行服务之前对 Server 的额外操作
	//
	// 比如添加模块等。不可以为空。
	Init func(*Server) error

	// 触发退出的信号
	//
	// 为空(nil 或是 []) 表示没有。
	Signals []os.Signal

	// 在初始化 Server 之前对 Options 的二次处理
	//
	// 可以为空。
	Options OptionsFunc

	// 命令行输出信息的通道
	//
	// 可以 os.Stdout 和 os.Stderr 选择，默认为 os.Stdout。
	Out io.Writer

	// 配置文件的加载器
	//
	// 为空则会给定一个能解析 .xml、.yaml、.yml 和 .json 文件的默认对象。
	Files *serialization.Files

	// 配置文件的文件名
	//
	// 相对于 CmdFS 参数指定的目录，默认为 web.xml。
	ConfigFilename string
}

// Exec 执行命令行操作
func (cmd *Command) Exec() error {
	if err := cmd.sanitize(); err != nil {
		panic(err) // Command 配置错误直接 panic
	}

	return cmd.exec()
}

func (cmd *Command) sanitize() error {
	if cmd.Name == "" {
		return errors.New("字段 Name 不能为空")
	}
	if cmd.Version == "" {
		return errors.New("字段 Version 不能为空")
	}
	if cmd.Init == nil {
		return errors.New("字段 Init 不能为空")
	}

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	if cmd.Files == nil {
		f := serialization.NewFiles(5)

		if err := f.Add(json.Marshal, json.Unmarshal, ".json"); err != nil {
			return err
		}

		if err := f.Add(xml.Marshal, xml.Unmarshal, ".xml"); err != nil {
			return err
		}

		if err := f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"); err != nil {
			return err
		}

		cmd.Files = f
	}

	if cmd.ConfigFilename == "" {
		cmd.ConfigFilename = "web.xml"
	}

	return nil
}

func (cmd *Command) exec() error {
	cl := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	cl.SetOutput(cmd.Out)

	v := cl.Bool("v", false, "显示版本号")
	h := cl.Bool("h", false, "显示帮助信息")
	f := cl.String("f", "./", "可读取的目录")
	i := cl.String("i", "", "执行的安装脚本")
	s := cl.Bool("s", true, "是否运行服务")

	if err := cl.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *v {
		_, err := fmt.Fprintln(cmd.Out, cmd.Name, cmd.Version)
		return err
	}

	if *h {
		cl.PrintDefaults()
		return nil
	}

	srv, err := LoadServer(cmd.Name, cmd.Version, cmd.Files, os.DirFS(*f), cmd.ConfigFilename, cmd.Options)
	if err != nil {
		return err
	}

	if err := cmd.Init(srv); err != nil {
		return err
	}

	if !*s { // 非服务
		return srv.Install(*i)
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
