// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/scheduled"
)

type (
	ScheduledJobFunc = scheduled.JobFunc
	ScheduledJob     = scheduled.Job
	Scheduler        = scheduled.Scheduler
)

// AddCron 添加新的定时任务
func (mgr *Manager) AddCron(title string, f ScheduledJobFunc, spec string, delay bool) {
	mgr.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
func (mgr *Manager) AddTicker(title string, f ScheduledJobFunc, dur time.Duration, imm, delay bool) {
	mgr.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
func (mgr *Manager) AddAt(title string, f ScheduledJobFunc, t time.Time, delay bool) {
	mgr.scheduled.At(title, f, t, delay)
}

// AddJob 添加新的计划任务
func (mgr *Manager) AddJob(title string, f ScheduledJobFunc, scheduler Scheduler, delay bool) {
	mgr.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回计划任务列表
func (mgr *Manager) Jobs() []*ScheduledJob { return mgr.scheduled.Jobs() }

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
