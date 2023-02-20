// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/localeutil"
)

type Services interface {
	// AddCron 添加新的定时任务
	//
	// f 表示服务的运行函数；
	// title 是对该服务的简要说明；
	// spec cron 表达式，支持秒；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	AddCron(title string, f JobFunc, spec string, delay bool)

	// AddTicker 添加新的定时任务
	//
	// f 表示服务的运行函数；
	// title 是对该服务的简要说明；
	// dur 时间间隔；
	// imm 是否立即执行一次该任务；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool)

	// AddAt 添加新的定时任务
	//
	// f 表示服务的运行函数；
	// title 是对该服务的简要说明；
	// t 指定的时间点；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	AddAt(title string, f JobFunc, ti time.Time, delay bool)

	// AddJob 添加新的计划任务
	//
	// f 表示服务的运行函数；
	// title 是对该服务的简要说明；
	// scheduler 计划任务的时间调度算法实现；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	AddJob(title string, f JobFunc, scheduler Scheduler, delay bool)

	// Add 添加新的服务
	//
	// f 表示服务的运行函数；
	// title 是对该服务的简要说明。
	//
	// NOTE: 如果服务已经处于运行的状态，则会自动运行新添加的服务。
	Add(title localeutil.LocaleStringer, f Servicer)

	// 添加新的服务
	AddFunc(title localeutil.LocaleStringer, f func(context.Context) error)

	// 获取所有的服务列表
	Services() []*Service

	// Jobs 返回所有的计划任务
	Jobs() []*Job
}
