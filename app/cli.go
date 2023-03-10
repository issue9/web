// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
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

// CLIOf 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//
//	-v 显示版本号；
//	-h 显示帮助信息；
//	-f 指定当前程序可读取的文件系统，这最终会转换成 [server.Server.FS]；
//	-a 执行的指令，该值会传递给 [CLIOf.Init]，由用户根据此值决定初始化方式；
//
// T 表示的是配置文件中的用户自定义数据类型。
type CLIOf[T any] struct {
	// NOTE: CLIOf 仅用于初始化 server.Server。对于接口的开发应当是透明的，
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

	app    *App
	action string
	fsys   fs.FS
}

// Exec 根据配置运行服务
//
// args 表示命令行参数，一般为 [os.Args]。
//
// 如果是 [CLIOf] 本身字段设置有问题会直接 panic，其它错误则返回该错误信息。
func (cmd *CLIOf[T]) Exec(args []string) (err error) {
	if err = cmd.sanitize(); err != nil { // 字段值有问题，直接 panic。
		panic(err)
	}

	f, do := cmd.FlagSet(true)
	f.SetOutput(cmd.Out)
	if err = f.Parse(args[1:]); err == nil {
		err = do(cmd.Out)
	}

	if err != nil {
		if le, ok := err.(localeutil.LocaleStringer); ok { // 对错误信息进行本地化转换
			err = errors.New(le.LocaleString(cmd.printer))
		}
	}
	return err
}

func (cmd *CLIOf[T]) sanitize() error {
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

	cmd.app = &App{
		ShutdownTimeout: cmd.ShutdownTimeout,
		Before: func() error {
			return CheckConfigSyntax[T](cmd.fsys, cmd.ConfigFilename)
		},
		NewServer: cmd.initServer,
	}

	return nil
}

// FlagSet 将当前对象转换成 [flag.FlagSet] 对象
//
// helpFlag 是否添加帮助选项，如果 FlagSet 是独立使用的，建议设置为 true，
// 或者直接使用 [CLIOf.Exec]，作为其它对象的子命令使用，可以设置为 false。
//
// fs 表示已经添加完参数的 FlagSet 对象，需要用户手动调用 [flag.FlagSet.Parse] 方法，
// 参数第一个元素应该是程序名称，如果是作为子命令使用，那么第一个元素应该是子命令的名称；
// do 表示实际执行的方法，其签名为 `func(w io.Writer) error`，w 表示处理过程中的输出通道。
func (cmd *CLIOf[T]) FlagSet(helpFlag bool) (fs *flag.FlagSet, do func(io.Writer) error) {
	flags := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	v := flags.Bool("v", false, cmd.printer.Sprintf("cmd.show_version"))
	f := flags.String("f", "./", cmd.printer.Sprintf("cmd.set_file_system"))
	flags.StringVar(&cmd.action, "a", "", cmd.printer.Sprintf("cmd.action"))

	var h bool
	if helpFlag {
		flags.BoolVar(&h, "h", false, cmd.printer.Sprintf("cmd.show_help"))
	}

	// 以上完成了 flag 的初始化

	return flags, func(w io.Writer) error {
		cmd.fsys = os.DirFS(*f)

		if *v {
			_, err := fmt.Fprintln(w, cmd.Name, cmd.Version)
			return err
		}

		if helpFlag && h {
			flags.PrintDefaults()
			return nil
		}

		if !sliceutil.Exists(cmd.ServeActions, func(e string) bool { return e == cmd.action }) { // 非服务
			_, err := cmd.initServer()
			return err
		}

		return cmd.app.Exec()
	}
}

// Restart 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成。
func (cmd *CLIOf[T]) Restart() { cmd.app.Restart() }

func (cmd *CLIOf[T]) initServer() (*server.Server, error) {
	srv, user, err := NewServerOf[T](cmd.Name, cmd.Version, cmd.fsys, cmd.ConfigFilename)
	if err != nil {
		return nil, errs.NewStackError(err)
	}

	if err = cmd.Init(srv, user, cmd.action); err != nil {
		return nil, errs.NewStackError(err)
	}

	return srv, nil
}

// CheckConfigSyntax 检测配置项语法是否正确
func CheckConfigSyntax[T any](fsys fs.FS, filename string) error {
	_, err := loadConfigOf[T](fsys, filename)
	return err
}
