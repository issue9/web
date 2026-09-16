// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kardianos/service"
	"golang.org/x/text/message"

	"github.com/issue9/cmdopt"
	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/config"
)

const (
	cmdVersion    = web.StringPhrase("cmd.version")
	cmdTestSyntax = web.StringPhrase("cmd.test_syntax")
	cmdUsage      = web.StringPhrase("cmd.usage")

	cmdDaemon      = web.StringPhrase("cmd.daemon")
	cmdDaemonUsage = web.StringPhrase("cmd.daemon_usage")

	cmdServe      = web.StringPhrase("cmd.serve")
	cmdServeUsage = web.StringPhrase("cmd.serve_usage")

	cmdHelp      = web.StringPhrase("cmd.help")
	cmdHelpUsage = web.StringPhrase("cmd.help_usage")
)

type CLIOptions[T comparable] struct {
	ID string // 程序 ID

	// 初始化 [web.Server]
	//
	// id, version 即为 [CLIOptions.ID] 和 [CLIOptions.Version]；
	// o 和 user 为从配置文件加载的数据信息；
	NewServer func(id, version string, o *server.Options, user T) (web.Server, error)

	// 其它子命令
	//
	// NOTE: 子命令名称不能是 daemon、help 和 serve
	Commands []*cmdopt.Command

	// 程序版本
	//
	// 如果为空，则会尝试调用 [web.GetAppVersion] 获得相关的值。
	Version string

	// 命令行输出信息的通道
	//
	// 默认为 [os.Stdout]。
	Out io.Writer

	// 配置文件所在的目录
	//
	// 具体参数说明可参考 [config.Load] 中 configDir 参数的说明。
	ConfigDir string

	// 配置文件的文件名
	//
	// 具体参数说明可参考 [config.Load] 中 filename 参数的说明。
	ConfigFilename string

	// 本地化的打印对象
	//
	// 若为空，则以 config.NewPrinter("*.yaml", locales.Locales) 进行初始化。
	//
	// 若是自定义，至少需要保证以下几个字符串的翻译项：
	//  - cmd.version
	//  - cmd.help
	//  - cmd.help_usage
	//  - can not be empty
	//  - cmd.test_syntax
	//  - cmd.usage
	//  - cmd.daemon
	//  - cmd.daemon_usage
	//  - cmd.serve
	//  - cmd.serve_usage
	//  - daemon status %s
	//  - syntax OK
	//  - invalid daemon control action: %s
	//  - command %s not found
	//  - [DaemonConfig.DisplayName]
	//  - [DaemonConfig.Description]
	//
	// NOTE: 此设置仅影响命令行的本地化，[web.Server] 的本地化由其自身管理。
	Printer *message.Printer

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	// 在命令行解析出错时的处理方式
	//
	// 默认值为 [flag.ContinueOnError]。
	ErrorHandling flag.ErrorHandling

	// 守护进程的设置项
	//
	// 如果该值不为 nil，将会注册 daemon 子命令，提供用以将当前应用转换为守护进程的相关操作。
	Daemon *DaemonConfig
	daemon *service.Config
}

type cli[T comparable] struct {
	App
	exec func(args []string) error
}

// NewCLI 提供一种简单的命令行生成方式
//
// 生成的命令行带以下参数和子命令：
//   - -v 显示版本号；
//   - -t 测试配置文件的语法是否正确；
//   - help 帮助子命令；
//   - daemon 守护进程子命令；
//   - serve 启动服务子命令；
//
// T 表示的是配置文件中的用户自定义数据类型，可参考 [config.Load] 中有关 User 的说明。
//
// 如果是 [CLIOptions] 本身字段设置有问题会直接 panic。
func NewCLI[T comparable](o *CLIOptions[T]) App {
	if err := o.sanitize(); err != nil { // 字段值有问题，直接 panic。
		panic(localeError(err, o.Printer))
	}

	app := newApp(o.ShutdownTimeout, func() (web.Server, error) {
		opt, user, err := config.Load[T](o.ConfigDir, o.ConfigFilename)
		if err != nil {
			return nil, web.NewStackError(err)
		}
		return o.NewServer(o.ID, o.Version, opt, user)
	})

	return &cli[T]{
		App: app,
		exec: func(args []string) error {
			rootCmd := func(fs *flag.FlagSet) cmdopt.DoFunc {
				v := fs.Bool("v", false, cmdVersion.LocaleString(o.Printer))
				t := fs.Bool("t", false, cmdTestSyntax.LocaleString(o.Printer))

				return func(w io.Writer) error {
					if *v {
						_, err := fmt.Fprintln(o.Out, o.ID, o.Version)
						return web.NewStackError(localeError(err, o.Printer))
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
							fmt.Fprintln(o.Out, web.StringPhrase("syntax OK").LocaleString(o.Printer))
						}
						return nil
					}

					fs.Usage()
					return nil
				}
			}

			opt := cmdopt.New(&cmdopt.Options{
				Name:          o.ID,
				Version:       o.Version,
				Output:        o.Out,
				ErrorHandling: o.ErrorHandling,
				UsageTemplate: cmdUsage.LocaleString(o.Printer),
				Command:       rootCmd,
				NotFound:      func(s string) string { return web.Phrase("command %s not found", s).LocaleString(o.Printer) },
			})

			// help 子命令
			cmdopt.Help(opt, "help", cmdHelp.LocaleString(o.Printer), cmdHelpUsage.LocaleString(o.Printer))

			// daemon 子命令
			if o.daemon != nil {
				opt.New("daemon", cmdDaemon.LocaleString(o.Printer), cmdDaemonUsage.LocaleString(o.Printer), func(fs *flag.FlagSet) cmdopt.DoFunc {
					return func(w io.Writer) error {
						a := "status"
						if fs.NArg() > 0 {
							a = fs.Arg(0)
						}

						status, err := app.runDaemon(a, o.daemon)
						if err == nil {
							_, err = fmt.Fprintln(w, web.Phrase("daemon status %s", statusString(status)).LocaleString(o.Printer))
						}
						return err
					}
				})
			}

			// serve 启动 web 服务
			opt.New("serve", cmdServe.LocaleString(o.Printer), cmdServeUsage.LocaleString(o.Printer), func(fs *flag.FlagSet) cmdopt.DoFunc {
				return func(w io.Writer) error { return localeError(app.Exec(), o.Printer) }
			})

			opt.NewCommand(o.Commands...)

			return opt.Exec(args[1:])
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
		o.Version = web.GetAppVersion("")
		if o.Version == "" {
			return web.NewFieldError("Version", locales.ErrCanNotBeEmpty())
		}
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
		o.daemon = o.Daemon.toServiceConfig(o.ID, o.Printer)
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
