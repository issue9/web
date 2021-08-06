// SPDX-License-Identifier: MIT

// Package cmd 命令行辅助工具
package cmd

import (
	"os"
	"os/signal"
	"time"

	"github.com/issue9/web/server"
)

type Command struct {
	s *server.Server
}

// New 返回 Command 对象
func New(s *server.Server, sig ...os.Signal) *Command {
	c := &Command{
		s: s,
	}
	if len(sig) > 0 {
		c.grace(sig...)
	}

	return c
}

func (c *Command) grace(sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := c.s.Close(3 * time.Second); err != nil {
			c.s.Logs().Error(err)
		}
		c.s.Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}

func (c *Command) Exec(tag string) error {
	if err := c.s.InitModules(tag); err != nil {
		return err
	}

	return c.s.Serve()
}
