// SPDX-License-Identifier: MIT

package service

import (
	"time"

	"github.com/issue9/scheduled"
)

type (
	JobFunc   = scheduled.JobFunc
	Job       = scheduled.Job
	Scheduler = scheduled.Scheduler
)

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddCron(title string, f JobFunc, spec string, delay bool) {
	srv.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool) {
	srv.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddAt(title string, f JobFunc, ti time.Time, delay bool) {
	srv.scheduled.At(title, f, ti, delay)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddJob(title string, f JobFunc, scheduler Scheduler, delay bool) {
	srv.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回所有的计划任务
func (srv *Server) Jobs() []*Job { return srv.scheduled.Jobs() }
