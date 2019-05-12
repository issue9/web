// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"time"

	"github.com/issue9/scheduled"
)

// JobFunc 定时任务执行的函数
type JobFunc = scheduled.JobFunc

// Job 定时任务实例
type Job = scheduled.Job

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f JobFunc, spec string, delay bool) {
	m.app.coreModule.AddInit(func() error {
		return m.app.scheduled.NewCron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f JobFunc, dur time.Duration, delay bool) {
	m.app.coreModule.AddInit(func() error {
		return m.app.scheduled.NewTicker(title, f, dur, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f JobFunc, spec string, delay bool) {
	m.app.coreModule.AddInit(func() error {
		return m.app.scheduled.NewAt(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// Schedulers 返回所有的计划任务
func (app *App) Schedulers() []*Job {
	return app.scheduled.Jobs()
}

func (app *App) scheduledService(ctx context.Context) error {
	go app.scheduled.Serve(app.logs.ERROR())

	<-ctx.Done()
	return context.Canceled
}
