// SPDX-License-Identifier: MIT

package web

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

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/localeutil"
	"github.com/issue9/web/config"
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
// NOTE: Command 的本地化信息采用当前用户的默认语言，
// 由 github.com/issue9/localeutil.DetectUserLanguageTag 读取。
type Command struct {
	Name string // 程序名称

	Version string // 程序版本

	// 在运行服务之前对 Server 的额外操作
	//
	// 比如添加模块等。不可以为空。
	Init func(s *Server, serve bool, install string) error

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
	// 默认为 os.Stdout。
	Out io.Writer

	// 配置文件的加载器
	//
	// 为空则会给定一个能解析 .xml、.yaml、.yml 和 .json 文件的默认对象。
	Files *Files

	// 配置文件的文件名
	//
	// 仅是文件名，相对的路径由命令行 -f 指定。
	// 如果为空，则直接采用 &Options{} 初始化 Server 对象。
	ConfigFilename string

	// 本地化的相关操作接口
	//
	// 如果为空，则会被初始化一个空对象。
	//
	// 当前命令行的翻译信息也由此对象提供，目前必须提供以下几个翻译项：
	//  -v  show version
	//  -h  show help
	//  -f  set file system
	//  -i  install key
	//  -s  run as server
	Catalog *catalog.Builder

	tag language.Tag
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

func (cmd *Command) exec() error {
	cl := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	cl.SetOutput(cmd.Out)
	p := message.NewPrinter(cmd.tag, message.Catalog(cmd.Catalog))

	v := cl.Bool("v", false, p.Sprintf("show version"))
	h := cl.Bool("h", false, p.Sprintf("show help"))
	f := cl.String("f", "./", p.Sprintf("set file system"))
	i := cl.String("i", "", p.Sprintf("install key"))
	s := cl.Bool("s", true, p.Sprintf("run as server"))

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

	opt, err := cmd.initOptions(os.DirFS(*f))
	if err != nil {
		return err
	}

	srv, err := NewServer(cmd.Name, cmd.Version, opt)
	if err != nil {
		return err
	}

	if err := cmd.Init(srv, *s, *i); err != nil {
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

func (cmd *Command) initOptions(fsys fs.FS) (opt *Options, err error) {
	if cmd.ConfigFilename != "" {
		opt, err = config.NewOptions(cmd.Files, fsys, cmd.ConfigFilename)
		if err != nil {
			return nil, err
		}
	} else {
		opt = &Options{
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
