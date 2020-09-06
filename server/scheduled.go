// SPDX-License-Identifier: MIT

package server

import (
	"context"

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
		if err := srv.scheduled.Serve(srv.Logs().ERROR(), srv.Logs().INFO()); err != nil {
			srv.Logs().Error(err)
		}
	}()

	<-ctx.Done()
	return context.Canceled
}
