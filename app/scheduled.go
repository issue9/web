// SPDX-License-Identifier: MIT

package app

import (
	"context"

	"github.com/issue9/scheduled"
)

// JobFunc 定时任务执行的函数
type JobFunc = scheduled.JobFunc

// Scheduled 获取 scheduled.Server 实例
func (app *App) Scheduled() *scheduled.Server {
	return app.scheduled
}

func (app *App) scheduledService(ctx context.Context) error {
	go func() {
		if err := app.scheduled.Serve(app.logs.ERROR(), app.logs.INFO()); err != nil {
			app.Logs().Error(err)
		}
	}()

	<-ctx.Done()
	return context.Canceled
}
