// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

// CLIOf 提供一种简单的命令行生成方式
//
// 生成的命令行带以下几个参数：
//
//	-v 显示版本号；
//	-h 显示帮助信息；
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

	// 配置文件所在的目录
	//
	// 这也将影响后续 [server.Options.Config] 变量，如果为空，则会采用 [server.DefaultConfigDir]。
	//
	// 有以下几种前缀用于指定不同的保存目录：
	//  - ~ 表示系统提供的配置文件目录，比如 Linux 的 XDG_CONFIG、Window 的 AppData 等；
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
	// 相对于 ConfigDir 的文件名。如果为空，表示不采用配置文件，
	// 由一个空的 [server.Options] 初始化对象，具体可以查看 [NewServerOf] 的实现。
	//
	// 需要保证序列化方法已经由 [RegisterFileSerializer] 注册；
	ConfigFilename string

	// 本地化的相关设置
	//
	// 若为空，则以 NewPrinter(locales.Locales, "*.yml") 进行初始化。
	//
	// NOTE: 此设置仅影响命令行的本地化(panic 信息不支持本地化)，[server.Server] 的本地化由其自身管理。
	Printer *localeutil.Printer

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	app    *App
	action string
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

	f := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	f.SetOutput(cmd.Out)
	do := cmd.FlagSet(true, f)
	if err = f.Parse(args[1:]); err == nil {
		err = do(cmd.Out)
	}

	if err != nil {
		if le, ok := err.(localeutil.LocaleStringer); ok { // 对错误信息进行本地化转换
			err = errors.New(le.LocaleString(cmd.Printer))
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

	if cmd.ConfigDir == "" {
		cmd.ConfigDir = server.DefaultConfigDir
	}

	if cmd.Printer == nil {
		cmd.Printer = NewPrinter(locales.Locales, "*.yml")
	}

	if cmd.Out == nil {
		cmd.Out = os.Stdout
	}

	cmd.app = &App{
		ShutdownTimeout: cmd.ShutdownTimeout,
		Before: func() error {
			return CheckConfigSyntax[T](cmd.ConfigDir, cmd.ConfigFilename)
		},
		NewServer: cmd.initServer,
	}

	return nil
}

const (
	cmdShowVersion = localeutil.StringPhrase("cmd.show_version")
	cmdAction      = localeutil.StringPhrase("cmd.action")
	cmdShowHelp    = localeutil.StringPhrase("cmd.show_help")
)

// FlagSet 将当前对象与 [flag.FlagSet] 关联
//
// helpFlag 是否添加帮助选项。
// 如果是独立使用的，建议设置为 true，或者直接使用 [CLIOf.Exec]。
// 作为子命令使用，可以设置为 false。
// fs 用于接收命令行的参数。
//
// do 表示实际执行的方法，其签名为 `func(w io.Writer) error`，w 表示处理过程中的输出通道。
func (cmd *CLIOf[T]) FlagSet(helpFlag bool, fs *flag.FlagSet) (do func(io.Writer) error) {
	v := fs.Bool("v", false, cmdShowVersion.LocaleString(cmd.Printer))
	fs.StringVar(&cmd.action, "a", "", cmdAction.LocaleString(cmd.Printer))

	var h bool
	if helpFlag {
		fs.BoolVar(&h, "h", false, cmdShowHelp.LocaleString(cmd.Printer))
	}

	return func(w io.Writer) error {
		if *v {
			_, err := fmt.Fprintln(w, cmd.Name, cmd.Version)
			return err
		}

		if helpFlag && h {
			fs.PrintDefaults()
			return nil
		}

		if !sliceutil.Exists(cmd.ServeActions, func(e string, _ int) bool { return e == cmd.action }) { // 非服务
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
	srv, user, err := NewServerOf[T](cmd.Name, cmd.Version, cmd.ConfigDir, cmd.ConfigFilename)
	if err != nil {
		return nil, errs.NewStackError(err)
	}

	if err = cmd.Init(srv, user, cmd.action); err != nil {
		return nil, errs.NewStackError(err)
	}

	return srv, nil
}

// CheckConfigSyntax 检测配置项语法是否正确
func CheckConfigSyntax[T any](configDir, filename string) error {
	_, err := loadConfigOf[T](configDir, filename)
	return err
}

// NewPrinter 根据参数构建一个本地化的打印对象
//
// 语言由 [localeutil.DetectUserLanguageTag] 决定。
// 参数指定了本地化的文件内容。
func NewPrinter(fsys fs.FS, glob string) *localeutil.Printer {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		log.Println(err) // 输出错误，但是不中断执行
	}

	s := make(config.Serializer, len(filesFactory))
	for k, v := range filesFactory {
		s.Add(v.Marshal, v.Unmarshal, k)
	}

	b := catalog.NewBuilder(catalog.Fallback(tag))
	locale.Load(s, b, fsys, glob)
	return message.NewPrinter(tag, message.Catalog(b))
}
