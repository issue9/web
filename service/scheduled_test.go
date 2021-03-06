// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
)

func TestScheduled(t *testing.T) {
	a := assert.New(t)

	mgr := NewManager(logs.New(), time.Local)
	a.Equal(0, len(mgr.Jobs()))

	err := mgr.AddAt("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.NotError(err)
	a.Equal(1, len(mgr.Jobs()))
}
