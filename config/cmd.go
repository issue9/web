// SPDX-License-Identifier: MIT

package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

// Command 提供一种简单的命令行方式
type Command struct {
	Out           io.Writer
	Name          string
	Version       string
	ServeTag      string
	ResultBuilder content.BuildResultFunc
	Signals       []os.Signal
}

// Exec 执行命令行操作
func (cmd *Command) Exec() {
	if err := cmd.exec(); err != nil {
		panic(err)
	}
}

func (cmd *Command) exec() error {
	v := flag.Bool("v", false, "显示版本号")
	tag := flag.String("tag", "", "执行的标签")
	f := flag.String("fs", "./", "可读取的目录")
	dir := os.DirFS(*f)

	if *v {
		fmt.Fprintln(cmd.Out, cmd.Name, cmd.Version)
		return nil
	}

	srv, err := NewServer(cmd.Name, cmd.Version, dir, cmd.ResultBuilder)
	if err != nil {
		return err
	}

	if err := srv.InitModules(*tag); err != nil {
		return err
	}
	if *tag == cmd.ServeTag {
		if cmd.Signals != nil { // nil 和 [] 处理方式不同，[] 表示所有信息

			cmd.grace(srv, cmd.Signals...)
		}
		return srv.Serve()
	}
	return nil
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
