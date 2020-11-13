// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"
)

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (mgr *Manager) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) error {
	return mgr.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (mgr *Manager) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) error {
	return mgr.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (mgr *Manager) AddAt(title string, f scheduled.JobFunc, spec string, delay bool) error {
	return mgr.scheduled.At(title, f, spec, delay)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (mgr *Manager) AddJob(title string, f scheduled.JobFunc, scheduler schedulers.Scheduler, delay bool) {
	mgr.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回计划任务列表
func (mgr *Manager) Jobs() []*scheduled.Job {
	return mgr.scheduled.Jobs()
}

func (mgr *Manager) scheduledService(ctx context.Context) error {
	go func() {
		if err := mgr.scheduled.Serve(); err != nil {
			mgr.logs.Error(err)
		}
	}()

	<-ctx.Done()
	mgr.scheduled.Stop()
	return context.Canceled
}
