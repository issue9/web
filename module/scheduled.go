// SPDX-License-Identifier: MIT

package module

import (
	"context"

	"github.com/issue9/scheduled"
)

// JobFunc 定时任务执行的函数
type JobFunc = scheduled.JobFunc

// Scheduled 获取 scheduled.Server 实例
func (srv *Modules) Scheduled() *scheduled.Server {
	return srv.scheduled
}

func (srv *Modules) scheduledService(ctx context.Context) error {
	go func() {
		if err := srv.scheduled.Serve(srv.ctxServer.Logs().ERROR(), srv.ctxServer.Logs().INFO()); err != nil {
			srv.ctxServer.Logs().Error(err)
		}
	}()

	<-ctx.Done()
	return context.Canceled
}
