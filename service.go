// SPDX-License-Identifier: MIT

package web

import (
	"time"

	"github.com/issue9/scheduled"
	"github.com/issue9/scheduled/schedulers"

	"github.com/issue9/web/context/service"
)

type (
	// ServiceFunc 服务实际需要执行的函数
	ServiceFunc = service.Func

	// ServiceState 服务的状态值
	ServiceState = service.State

	// Service 服务模型
	Service = service.Service

	// JobFunc 计划任务执行的函数
	JobFunc = scheduled.JobFunc

	// Job 计划任务的模型
	Job = scheduled.Job

	// Scheduler 计划任务的调试方法需要实现的接口
	Scheduler = schedulers.Scheduler
)

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (m *Module) AddService(f ServiceFunc, title string) {
	m.AddInit(func() error {
		m.web.CTXServer().Services().AddService(f, title)
		return nil
	}, "注册服务："+title)
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddCron(title string, f JobFunc, spec string, delay bool) {
	m.AddInit(func() error {
		return m.web.CTXServer().Services().AddCron(title, f, spec, delay)
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
		return m.web.CTXServer().Services().AddTicker(title, f, dur, imm, delay)
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
		return m.web.CTXServer().Services().AddAt(title, f, spec, delay)
	}, "注册计划任务"+title)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (m *Module) AddJob(title string, f JobFunc, scheduler Scheduler, delay bool) {
	m.AddInit(func() error {
		m.web.CTXServer().Services().AddJob(title, f, scheduler, delay)
		return nil
	}, "注册计划任务"+title)
}
