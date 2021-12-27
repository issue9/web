// SPDX-License-Identifier: MIT

// Package app 提供可选的初始化程序的方法
//
// NOTE: app 并不是必须的，只是为用户提供了一种简便的方式构建程序，
// 相对地也会有诸多限制，如果觉得不适用，可以直接使用 server.New 下的内容。
package app

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// App 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//  - v 显示版本号；
//  - h 显示帮助信息；
//  - f 指定当前程序可读取的文件系统，这最终会转换成 Server.FS；
//  - a 执行的动作，该值会传递给 Init，由用户根据 a 初始化方式；
//  - s 以服务的形式运行；
//
// 本地化信息采用当前用户的默认语言，
// 由 github.com/issue9/localeutil.DetectUserLanguageTag 决定。
// 如果想让 App 支持本地化操作，最起码需要向 Catalog 注册命令行参数的本地化信息：
//  -v  show version
//  -h  show help
//  -f  set file system
//  -a  action
//  -s  run as server
// 对于 App 的初始化错误产生的 panic 信息是不支持本地的。
//
//  // 本地化命令行的帮助信息
//  builder := catalog.NewBuilder()
//  builder.SetString("show help", "显示帮助信息")
//  builder.SetString("show version", "显示版本信息")
//
//  app := &web.App{
//      Name: "app",
//      Version: "1.0.0",
//      Init: func(s *Server) error {...},
//      Catalog: builder,
//  }
//
//  app.Exec()
type App struct {
	Name string // 程序名称

	Version string // 程序版本

	// 在运行服务之前对 Server 的额外操作
	//
	// 比如添加模块等。不可以为空。
	Init func(s *server.Server, action string) error

	// 在初始化 Server 之前对 Options 的二次处理
	//
	// 可以为空。
	Options func(*server.Options)

	// 命令行输出信息的通道
	//
	// 默认为 os.Stdout。
	Out io.Writer

	// 配置文件的加载器
	//
	// 为空则会给定一个能解析 .xml、.yaml、.yml 和 .json 文件的默认对象。
	// 该值可能会被 Options 操作所覆盖。
	Files *serialization.Files

	// 配置文件的文件名
	//
	// 仅是文件名，相对的路径由命令行 -f 指定。
	// 如果为空，则直接采用 &Options{} 初始化 Server 对象。
	ConfigFilename string

	// 本地化的相关操作接口
	//
	// 如果为空，则会被初始化一个空对象，该值可能会被 Options 操作所覆盖。
	Catalog *catalog.Builder

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
func (cmd *App) Exec(args []string) error {
	if err := cmd.sanitize(); err != nil {
		panic(err) // Command 配置错误直接 panic
	}
	return cmd.exec(args)
}

func (cmd *App) sanitize() error {
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

	return nil
}

func (cmd *App) exec(args []string) error {
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

	opt, err := cmd.initOptions(os.DirFS(*f))
	if err != nil {
		return err
	}

	srv, err := server.New(cmd.Name, cmd.Version, opt)
	if err != nil {
		return err
	}

	if err := cmd.Init(srv, *a); err != nil {
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

func (cmd *App) initOptions(fsys fs.FS) (opt *server.Options, err error) {
	if cmd.ConfigFilename != "" {
		opt, err = NewOptions(cmd.Files, fsys, cmd.ConfigFilename)
		if err != nil {
			return nil, err
		}
	} else {
		opt = &server.Options{
			FS:    fsys,
			Files: cmd.Files,
		}
	}
	opt.Catalog = cmd.Catalog

	if cmd.Options != nil {
		cmd.Options(opt)
	}

	return opt, nil
}

func (cmd *App) grace(s *server.Server, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := s.Close(cmd.SignalTimeout); err != nil {
			s.Logs().Error(err)
		}
		s.Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}
