// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"time"

	"github.com/issue9/scheduled"
)

// JobFunc 定时任务执行的函数
type JobFunc = scheduled.JobFunc

// Job 描述计划任务的信息。
type Job struct {
	State     scheduled.State
	Title     string
	Prev      time.Time
	Next      time.Time
	Scheduled string
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f JobFunc, spec string, delay bool) {
	m.ms.coreModule.AddInit(func() error {
		return m.ms.scheduled.NewCron(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddTicker(title string, f JobFunc, dur time.Duration, delay bool) {
	m.ms.coreModule.AddInit(func() error {
		return m.ms.scheduled.NewTicker(title, f, dur, delay)
	}, "注册计划任务"+title)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddAt(title string, f JobFunc, spec string, delay bool) {
	m.ms.coreModule.AddInit(func() error {
		return m.ms.scheduled.NewAt(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// Jobs 返回所有的计划任务
func (ms *Modules) Jobs() []*Job {
	jobs := ms.scheduled.Jobs()

	ret := make([]*Job, 0, len(jobs))
	for _, job := range jobs {
		ret = append(ret, &Job{
			State:     job.State(),
			Title:     job.Name(),
			Prev:      job.Prev(),
			Next:      job.Next(),
			Scheduled: job.Scheduler.Title(),
		})
	}

	return ret
}

func (ms *Modules) scheduledService(ctx context.Context) error {
	go ms.scheduled.Serve(ms.logs.ERROR())

	<-ctx.Done()
	return context.Canceled
}
