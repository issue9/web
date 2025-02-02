// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/kardianos/service"
	"golang.org/x/text/message"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/config"
)

const (
	cmdShowVersion = web.StringPhrase("cmd.show_version")
	cmdAction      = web.StringPhrase("cmd.action")
	cmdDaemon      = web.StringPhrase("cmd.daemon")
	cmdShowHelp    = web.StringPhrase("cmd.show_help")
	cmdTestSyntax  = web.StringPhrase("cmd.test_syntax")
)

type CLIOptions[T comparable] struct {
	ID      string // 程序 ID
	Version string // 程序版本

	// 初始化 [web.Server]
	//
	// id, version 即为 [CLIOptions.ID] 和 [CLIOptions.Version]；
	// o 和 user 为从配置文件加载的数据信息；
	// action 为 -a 命令行指定的参数；
	NewServer func(id, version string, o *server.Options, user T, action string) (web.Server, error)

	// 以服务运行的指令
	ServeActions []string

	// 命令行输出信息的通道
	//
	// 默认为 [os.Stdout]。
	Out io.Writer

	// 配置文件所在的目录
	//
	// 有以下几种前缀用于指定不同的保存目录：
	//  - ~ 表示系统提供的配置文件目录，比如 Linux 的 XDG_CONFIG、Windows 的 AppData 等；
	//  - @ 表示当前程序的主目录；
	//  - ^ 表示绝对路径；
	//  - # 表示工作路径；
	//  - 其它则是直接采用 [config.Dir] 初始化。
	// 如果为空则采用 [server.DefaultConfigDir] 中指定的值。
	//
	// NOTE: 具体说明可参考 [config.BuildDir] 的 dir 参数。
	ConfigDir string

	// 配置文件的文件名
	//
	// 相对于 ConfigDir 的文件名，不能为空。
	//
	// 需要保证序列化方法已经由 [config.RegisterFileSerializer] 注册；
	ConfigFilename string

	// 本地化的打印对象
	//
	// 若为空，则以 config.NewPrinter("*.yaml", locales.Locales) 进行初始化。
	//
	// 若是自定义，至少需要保证以下几个字符串的翻译项，才有效果：
	//  - cmd.show_version
	//  - cmd.action
	//  - cmd.show_help
	//  - can not be empty
	//  - cmd.test_syntax
	//  - status %v
	//  - syntax OK
	//
	// NOTE: 此设置仅影响命令行的本地化，[web.Server] 的本地化由其自身管理。
	Printer *message.Printer

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	// 在命令行解析出错时的处理方式
	//
	// 默认值为 [flag.ContinueOnError]
	ErrorHandling flag.ErrorHandling

	// 守护进程的设置项
	//
	// 如果该值不为 nil，将会注册 -d 选项，
	// 通过 -d 可以对服务进行以下操作：
	//  - install 安装服务
	//  - uninstall 卸载服务
	//  - start 启动服务
	//  - stop 停止服务
	//  - restart 重启服务
	//  - status 查看服务状态
	Daemon *service.Config
}

type cli[T comparable] struct {
	App
	exec func(args []string) error
}

// NewCLI 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//   - -v 显示版本号；
//   - -h 显示帮助信息；
//   - -a 执行的指令，该值会传递给 [CLIOptions.NewServer]，由用户根据此值决定初始化方式；
//   - -d 将当前程序作为守护进程的相关操作；
//   - -t 测试配置文件的语法是否正确；
//
// T 表示的是配置文件中的用户自定义数据类型，可参考 [config.Load] 中有关 User 的说明。
//
// 如果是 [CLIOptions] 本身字段设置有问题会直接 panic。
func NewCLI[T comparable](o *CLIOptions[T]) App {
	if err := o.sanitize(); err != nil { // 字段值有问题，直接 panic。
		panic(localeError(err, o.Printer))
	}

	var action string // -a 参数

	initServer := func() (web.Server, error) {
		opt, user, err := config.Load[T](o.ConfigDir, o.ConfigFilename)
		if err != nil {
			return nil, web.NewStackError(err)
		}
		return o.NewServer(o.ID, o.Version, opt, user, action)
	}

	app := newApp(o.ShutdownTimeout, initServer)

	return &cli[T]{
		App: app,
		exec: func(args []string) (err error) {
			fs := flag.NewFlagSet(o.ID, o.ErrorHandling)
			fs.SetOutput(o.Out)

			v := fs.Bool("v", false, cmdShowVersion.LocaleString(o.Printer))
			h := fs.Bool("h", false, cmdShowHelp.LocaleString(o.Printer))
			t := fs.Bool("t", false, cmdTestSyntax.LocaleString(o.Printer))
			fs.StringVar(&action, "a", "", cmdAction.LocaleString(o.Printer))

			var daemon string // -d 选项
			if o.Daemon != nil {
				fs.StringVar(&daemon, "d", "", cmdDaemon.LocaleString(o.Printer))
			}

			if err = fs.Parse(args[1:]); err != nil {
				return web.NewStackError(localeError(err, o.Printer))
			}

			if o.Daemon != nil && daemon != "" { // 在其它选项之前
				status, err := app.runDaemon(daemon, o.Daemon)
				if err != nil {
					return err
				}

				_, err = fmt.Fprintln(o.Out, o.Printer.Sprintf("status %v", status))
				return err
			}

			if *v {
				_, err = fmt.Fprintln(o.Out, o.ID, o.Version)
				return web.NewStackError(localeError(err, o.Printer))
			}

			if *h {
				fs.PrintDefaults()
				return nil
			}

			if *t {
				_, _, err := config.Load[T](o.ConfigDir, o.ConfigFilename)
				if err != nil {
					var msg string
					if le, ok := err.(web.LocaleStringer); ok { // 对错误信息进行本地化转换
						msg = le.LocaleString(o.Printer)
					} else {
						msg = err.Error()
					}

					fmt.Fprintln(o.Out, msg)
				} else {
					fmt.Fprintln(o.Out, web.Phrase("syntax OK").LocaleString(o.Printer))
				}
				return nil
			}

			if slices.Index(o.ServeActions, action) < 0 { // 非服务
				_, err = initServer()
				return localeError(err, o.Printer)
			}

			return localeError(app.Exec(), o.Printer)
		},
	}
}

func (cmd *cli[T]) Exec() error { return cmd.exec(os.Args) }

func (o *CLIOptions[T]) sanitize() error {
	if o.Printer == nil {
		p, err := config.NewPrinter("*.yaml", locales.Locales...)
		if err != nil {
			return err
		}
		o.Printer = p
	}

	if o.ID == "" {
		return web.NewFieldError("ID", locales.ErrCanNotBeEmpty())
	}
	if o.Version == "" {
		return web.NewFieldError("Version", locales.ErrCanNotBeEmpty())
	}
	if o.NewServer == nil {
		return web.NewFieldError("NewServer", locales.ErrCanNotBeEmpty())
	}

	if o.ConfigDir == "" {
		o.ConfigDir = server.DefaultConfigDir
	}

	if o.Out == nil {
		o.Out = os.Stdout
	}

	if o.ErrorHandling == 0 {
		o.ErrorHandling = flag.ContinueOnError
	}

	if o.Daemon != nil {
		if o.Daemon.Name == "" {
			o.Daemon.Name = o.ID
		}
	}

	return nil
}

func localeError(err error, p *message.Printer) error {
	if err != nil {
		if le, ok := err.(web.LocaleStringer); ok { // 对错误信息进行本地化转换
			return errors.New(le.LocaleString(p))
		}
	}
	return err
}
