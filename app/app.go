// SPDX-License-Identifier: MIT

// Package app 提供构建程序的简便方法
//
// app 并不是必须的，只是为用户提供了一种简便的方式构建程序，
// 相对地也会有诸多限制，如果觉得不适用，可以自行调用 [server.New]。
package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/server"
)

// AppOf 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//  - v 显示版本号；
//  - h 显示帮助信息；
//  - f 指定当前程序可读取的文件系统，这最终会转换成 Server.FS；
//  - a 执行的动作，该值会传递给 Init，由用户根据 a 决定初始化方式；
//  - s 以服务的形式运行；
//
// 本地化信息采用当前用户的默认语言，
// 由 github.com/issue9/localeutil.DetectUserLanguageTag 决定。
// 如果想让 AppOf 支持本地化操作，需要向 Catalog 注册命令行参数的本地化信息：
//  -v  show version
//  -h  show help
//  -f  set file system
//  -a  action
//  -s  run as server
// 对于 AppOf 的初始化错误产生的 panic 信息是不支持本地的。
//
//  // 本地化命令行的帮助信息
//  builder := catalog.NewBuilder()
//  builder.SetString("show help", "显示帮助信息")
//  builder.SetString("show version", "显示版本信息")
//
//  cmd := &app.AppOf[struct{}]{
//      Name: "app",
//      Version: "1.0.0",
//      Init: func(s *Server) error {...},
//      Catalog: builder,
//  }
//
//  cmd.Exec()
//
// T 表示的是配置文件中的用户自定义数据类型，如果不需要可以设置为 struct{}。
type AppOf[T any] struct {
	// NOTE: AppOf 仅用于初始化 server.Server，不应当赋予 AppOf 太多的功能。
	// AppOf 对于接口的开发应当是透明的，开发者所有的功能都可以通过 Context
	// 和 Server 获得。

	Name    string // 程序名称
	Version string // 程序版本

	// 在运行服务之前对 [server.Server] 的额外操作
	//
	// 比如添加模块等。不可以为空。
	// user 为用户自定义的数据类型；
	// action 为 -a 命令行指定的参数；
	Init func(s *server.Server, user *T, action string) error

	// 命令行输出信息的通道
	//
	// 默认为 os.Stdout。
	Out io.Writer

	// 配置文件的文件名
	//
	// 需要保证 [RegisterFileSerializer] 能解析此文件指定的内容；
	//
	// 仅是文件名，相对的路径由命令行 -f 指定。
	ConfigFilename string

	// 设置生成 [problem.Problem] 对象的方法
	//
	// 如果为空，则由 [server.Server.New] 决定其默认值。
	ProblemBuilder problem.BuildFunc

	// 本地化 AppOf 中的命令行信息
	//
	// 可以为空，那么这些命令行信息将显示默认内容。
	Catalog catalog.Catalog

	// 触发退出的信号
	//
	// 为空(nil 或是 []) 表示没有。
	Signals []os.Signal

	// 通过信号触发退出时的等待时间
	SignalTimeout time.Duration

	tag language.Tag
}

// Exec 执行命令行操作
//
// args 表示命令行参数，一般为 os.Args，采用明确的参数传递，方便测试用。
func (cmd *AppOf[T]) Exec(args []string) error {
	if err := cmd.sanitize(); err != nil {
		os.Exit(2)
	}
	return cmd.exec(args)
}

func (cmd *AppOf[T]) sanitize() error {
	if cmd.Name == "" {
		return errors.New("字段 Name 不能为空")
	}
	if cmd.Version == "" {
		return errors.New("字段 Version 不能为空")
	}
	if cmd.Init == nil {
		return errors.New("字段 Init 不能为空")
	}

	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		fmt.Println(err) // 输出错误，但是不中断执行
	}
	cmd.tag = tag

	if cmd.Catalog == nil {
		cmd.Catalog = catalog.NewBuilder(catalog.Fallback(cmd.tag))
	}

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	return nil
}

func (cmd *AppOf[T]) exec(args []string) error {
	cl := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	cl.SetOutput(cmd.Out)
	p := message.NewPrinter(cmd.tag, message.Catalog(cmd.Catalog))

	v := cl.Bool("v", false, p.Sprintf("show version"))
	h := cl.Bool("h", false, p.Sprintf("show help"))
	f := cl.String("f", "./", p.Sprintf("set file system"))
	a := cl.String("a", "", p.Sprintf("action"))
	s := cl.Bool("s", false, p.Sprintf("run as server"))

	if err := cl.Parse(args[1:]); err != nil {
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

	srv, user, err := NewServerOf[T](cmd.Name, cmd.Version, cmd.ProblemBuilder, os.DirFS(*f), cmd.ConfigFilename)
	if err != nil {
		return err
	}

	if err = cmd.Init(srv, user, *a); err != nil {
		return err
	}

	if !*s { // 非服务
		return nil
	}

	if len(cmd.Signals) > 0 {
		cmd.grace(srv, cmd.Signals...)
	}

	return srv.Serve()
}

func (cmd *AppOf[T]) grace(s *server.Server, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := s.Close(cmd.SignalTimeout); err != nil {
			io.WriteString(cmd.Out, err.Error())
		}
	}()
}
