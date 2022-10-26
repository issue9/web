// SPDX-License-Identifier: MIT

// Package app 为构建程序提供相对简便的方法
//
// app 并不是必须的，只是为用户提供了一种简便的方式构建程序，
// 相对地也会有诸多限制，如果觉得不适用，可以自行调用 [server.New]。
package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/server"
)

// AppOf 提供一种简单的命令行生成方式
//
// T 表示的是配置文件中的用户自定义数据类型，如果不需要可以设置为 struct{}。
//
// 生成的命令行带以下几个参数：
//
//	-v 显示版本号；
//	-h 显示帮助信息；
//	-f 指定当前程序可读取的文件系统，这最终会转换成 [server.Server.FS]；
//	-a 执行的动作，该值会传递给 Init，由用户根据 a 决定初始化方式；
//
// 通过向 [AppOf.Catalog] 注册本地化字符串，可以让命令行支持本地化显示：
//
//	// 构建 catalog.Catalog 实例
//	builder := catalog.NewBuilder()
//	builder.SetString("show help", "显示帮助信息")
//	builder.SetString("show version", "显示版本信息")
//
//	cmd := &app.AppOf[struct{}]{
//	    Name: "app",
//	    Version: "1.0.0",
//	    Init: func(s *Server) error {...},
//	    Catalog: builder,
//	}
//
//	cmd.Exec()
//
// 由 [localeutil.DetectUserLanguageTag] 检测当前系统环境并显示，
// 本地化命令行参数需要提供以下翻译项：
//
//	-show version
//	-show help
//	-set file system
//	-action
//	-[AppOf.Desc] 的内容
//
// 在支持 HUP 信号的系统，会接收 HUP 信号用于重启服务（调用 [AppOf.Restart]）。
//
// NOTE: panic 信息是不支持本地化。
type AppOf[T any] struct {
	// NOTE: AppOf 仅用于初始化 server.Server。对于接口的开发应当是透明的，
	// 开发者所有的功能都可以通过 Context 和 Server 获得。

	Name    string // 程序名称
	Version string // 程序版本
	Desc    string // 程序描述

	// 在运行服务之前对 [server.Server] 的额外操作
	//
	// 比如添加模块等。不可以为空。
	// user 为用户自定义的数据结构；
	// action 为 -a 命令行指定的参数；
	Init func(s *server.Server, user *T, action string) error

	// 以服务运行的指令
	ServeActions []string

	// 命令行输出信息的通道
	//
	// 默认为 [os.Stdout]。
	Out io.Writer

	// 配置文件的文件名
	//
	// 需要保证 [RegisterFileSerializer] 能解析此文件指定的内容；
	//
	// 仅是文件名，相对的路径由命令行 -f 指定。
	// 如果为空，表示不采用配置文件，由一个空的 [server.Options] 初始化对象，
	// 具体可以查看 [NewServerOf] 的实现。
	ConfigFilename string

	// 生成 [server.Problem] 对象的方法
	//
	// 如果为空，则由 [server.Options] 决定其默认值。
	ProblemBuilder server.BuildProblemFunc

	// 本地化 AppOf 中的命令行信息
	//
	// 如果为空，那么这些命令行信息将显示默认内容。
	Catalog catalog.Catalog

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	tag language.Tag

	// 重启服务的相关选项
	srv     *server.Server
	restart bool
	action  string
	fsys    fs.FS
}

// Exec 根据配置运行服务
//
// args 表示命令行参数，一般为 [os.Args]。
//
// 如果是 AppOf 本身字段设置有问题会直接 panic，其它错误则返回该错误信息。
func (cmd *AppOf[T]) Exec(args []string) error {
	if err := cmd.sanitize(); err != nil { // AppOf 字段值有问题，直接 panic。
		panic(err)
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
	flags := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	flags.SetOutput(cmd.Out)
	p := message.NewPrinter(cmd.tag, message.Catalog(cmd.Catalog))

	v := flags.Bool("v", false, p.Sprintf("show version"))
	h := flags.Bool("h", false, p.Sprintf("show help"))
	f := flags.String("f", "./", p.Sprintf("set file system"))
	flags.StringVar(&cmd.action, "a", "", p.Sprintf("action"))
	flags.Usage = func() {
		fmt.Fprintln(cmd.Out, p.Sprintf(cmd.Desc))
	}
	if err := flags.Parse(args[1:]); err != nil {
		return errs.StackError(err)
	}

	// 以上完成了 flag 的初始化

	cmd.fsys = os.DirFS(*f)

	if *v {
		_, err := fmt.Fprintln(cmd.Out, cmd.Name, cmd.Version)
		return err
	}

	if *h {
		flags.PrintDefaults()
		return nil
	}

	if !sliceutil.Exists(cmd.ServeActions, func(e string) bool { return e == cmd.action }) { // 非服务
		return cmd.initServer()
	}

	cmd.hup() // 注册 SIGHUP 信号

RESTART:
	if err := cmd.initServer(); err != nil {
		return err
	}

	cmd.restart = false
	err := cmd.srv.Serve()
	if cmd.restart { // 等待 Serve 过程中，如果调用 Restart，会将 cmd.restart 设置为 true。
		goto RESTART
	}
	return err
}

// Restart 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成，
// 也无法知晓新服务的状态。如果返回了错误信息，只能表示关闭旧服务时出错或是配置文件有问题。
//
// 此操作会重新加配置文件，如果配置文件有问题，将不会重启且返回错误信息。
func (cmd *AppOf[T]) Restart() error {
	cmd.restart = true

	if err := CheckConfigSyntax[T](cmd.fsys, cmd.ConfigFilename); err != nil {
		return err
	}

	return cmd.srv.Close(cmd.ShutdownTimeout)
}

func (cmd *AppOf[T]) initServer() error {
	srv, user, err := NewServerOf[T](cmd.Name, cmd.Version, cmd.ProblemBuilder, cmd.fsys, cmd.ConfigFilename)
	if err != nil {
		return errs.StackError(err)
	}

	if err = cmd.Init(srv, user, cmd.action); err != nil {
		return errs.StackError(err)
	}

	cmd.srv = srv
	return nil
}

func (cmd *AppOf[T]) hup() {
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGHUP)

		for range signalChannel {
			if err := cmd.Restart(); err != nil {
				fmt.Fprintln(cmd.Out, errs.StackError(err))
			}
		}
	}()
}

// CheckConfigSyntax 检测配置项语法是否正确
func CheckConfigSyntax[T any](fsys fs.FS, filename string) error {
	_, err := loadConfigOf[T](fsys, filename)
	return err
}
