// SPDX-License-Identifier: MIT

package module

import (
	"context"
	"time"

	"github.com/issue9/scheduled"
)

// JobFunc 定时任务执行的函数
type JobFunc = scheduled.JobFunc

// Scheduled 获取 scheduled.Server 实例
func (srv *Server) Scheduled() *scheduled.Server {
	return srv.scheduled
}

func (srv *Server) scheduledService(ctx context.Context) error {
	go func() {
		if err := srv.scheduled.Serve(srv.ctxServer.Logs().ERROR(), srv.ctxServer.Logs().INFO()); err != nil {
			srv.ctxServer.Logs().Error(err)
		}
	}()

	<-ctx.Done()
	return context.Canceled
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.srv.Scheduled().Cron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool) {
	m.AddInit(func() error {
		return m.srv.Scheduled().Tick(title, f, dur, imm, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.srv.Scheduled().At(title, f, spec, delay)
	}, "注册计划任务"+title)
}
