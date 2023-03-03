// SPDX-License-Identifier: MIT

// Package app 为构建程序提供相对简便的方法
//
// 提供了两种方式构建服务：
//   - [NewServerOf] 从配置文件构建 [server.Server] 对象；
//   - [AppOf] 直接生成一个简单的命令行程序；
//
// NOTE: 这并不一个必需的包，如果觉得不适用，可以直接采用 [server.New] 初始化服务。
//
// # 配置文件
//
// [NewServerOf] 和 [AppOf] 都是通过加载配置文件对项目进行初始化。
// 对于配置文件各个字段的定义，可参考源代码，入口在 config.go 文件的 configOf 对象。
// 配置文件中除了固定的字段之外，还提供了泛型变量 User 用于指定用户自定义的额外字段。
//
// # 注册函数
//
// 当前包提供大量的注册函数，以用将某些无法直接采用序列化的内容转换可序列化的。
// 比如通过 [RegisterEncoding] 将 `gzip-default` 等字符串表示成压缩算法，
// 以便在配置文件进行指定。
//
// 所有的注册函数处理逻辑上都相似，碰上同名的会覆盖，否则是添加。
// 且默认情况下都提供了一些可选项，只有在用户需要额外添加自己的内容时才需要调用注册函数。
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
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/files"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

// AppOf 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//
//	-v 显示版本号；
//	-h 显示帮助信息；
//	-f 指定当前程序可读取的文件系统，这最终会转换成 [server.Server.FS]；
//	-a 执行的指令，该值会传递给 [AppOf.Init]，由用户根据此值决定初始化方式；
//
// T 表示的是配置文件中的用户自定义数据类型。
type AppOf[T any] struct {
	// NOTE: AppOf 仅用于初始化 server.Server。对于接口的开发应当是透明的，
	// 开发者所有的功能都可以通过 Context 和 Server 获得。

	Name    string // 程序名称
	Version string // 程序版本

	// 在运行服务之前对 [server.Server] 的额外操作
	//
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
	// 仅是文件名，相对的路径由命令行 -f 指定。如果为空，表示不采用配置文件，
	// 由一个空的 [server.Options] 初始化对象，具体可以查看 [NewServerOf] 的实现。
	//
	// 需要保证序列化方法已经由 [RegisterFileSerializer] 注册；
	ConfigFilename string

	// 本地化的相关设置
	//
	// LocaleFS 本地化文件所在的文件系统，如果为空则指向 [locales.Locales]，
	// LocaleGlob 从 LocaleFS 中查找本地化文件的匹配模式，如果为空则为 *.yaml。
	// LocaleGlob 指定的文件格式必须是已经通过 [RegisterFileSerializer] 注册的。
	// 由 [localeutil.DetectUserLanguageTag] 检测当前系统环境并决定采用哪种语言。
	//
	// NOTE: 此设置仅影响命令行的本地化(panic 信息不支持本地化)，
	// [server.Server] 的本地化由其自身管理。
	LocaleFS   fs.FS
	LocaleGlob string
	printer    *message.Printer

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

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

	if err := cmd.exec(args); err != nil {
		if le, ok := err.(localeutil.LocaleStringer); ok { // 对错误信息进行本地化转换
			return errors.New(le.LocaleString(cmd.printer))
		}
		return err
	}
	return nil
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

	if cmd.LocaleFS == nil {
		cmd.LocaleFS = locales.Locales
	}
	if cmd.LocaleGlob == "" {
		cmd.LocaleGlob = "*.yaml"
	}

	b := catalog.NewBuilder(catalog.Fallback(tag))
	files.LoadLocales(buildFiles(cmd.LocaleFS), b, nil, cmd.LocaleGlob)
	cmd.printer = message.NewPrinter(tag, message.Catalog(b))

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	return nil
}

func (cmd *AppOf[T]) exec(args []string) error {
	flags := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	flags.SetOutput(cmd.Out)

	v := flags.Bool("v", false, cmd.printer.Sprintf("cmd.show_version"))
	h := flags.Bool("h", false, cmd.printer.Sprintf("cmd.show_help"))
	f := flags.String("f", "./", cmd.printer.Sprintf("cmd.set_file_system"))
	flags.StringVar(&cmd.action, "a", "", cmd.printer.Sprintf("cmd.action"))
	if err := flags.Parse(args[1:]); err != nil {
		return errs.NewStackError(err)
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
// 也无法知晓新服务的状态。如果返回了错误信息，
// 只能表示关闭旧服务时出错或是配置文件有语法错误。
func (cmd *AppOf[T]) Restart() error {
	cmd.restart = true

	if err := CheckConfigSyntax[T](cmd.fsys, cmd.ConfigFilename); err != nil {
		return err
	}

	return cmd.srv.Close(cmd.ShutdownTimeout)
}

func (cmd *AppOf[T]) initServer() error {
	srv, user, err := NewServerOf[T](cmd.Name, cmd.Version, cmd.fsys, cmd.ConfigFilename)
	if err != nil {
		return errs.NewStackError(err)
	}

	if err = cmd.Init(srv, user, cmd.action); err != nil {
		return errs.NewStackError(err)
	}

	cmd.srv = srv
	return nil
}

// CheckConfigSyntax 检测配置项语法是否正确
func CheckConfigSyntax[T any](fsys fs.FS, filename string) error {
	_, err := loadConfigOf[T](fsys, filename)
	return err
}

// SignalHUP 让 AppOf 支持 [HUP] 信号
//
// [HUP]: https://en.wikipedia.org/wiki/SIGHUP
func SignalHUP[T any](cmd *AppOf[T]) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)

	go func() {
		for range signalChannel {
			if err := cmd.Restart(); err != nil {
				fmt.Fprintln(cmd.Out, errs.NewStackError(err))
			}
		}
	}()
}
