// SPDX-License-Identifier: MIT

package service

import (
	"context"

	"github.com/issue9/scheduled"
)

func (mgr *Manager) Scheduled() *scheduled.Server { return mgr.scheduled }

func (mgr *Manager) scheduledService(ctx context.Context) error {
	go func() {
		if err := mgr.Scheduled().Serve(); err != nil {
			mgr.logs.Error(err)
		}
	}()

	<-ctx.Done()
	mgr.scheduled.Stop()
	return context.Canceled
}
