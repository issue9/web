// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert/v2"
)

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)
	a.Equal(0, len(srv.Jobs()))

	srv.scheduled.At("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(srv.scheduled.Jobs()))
}
