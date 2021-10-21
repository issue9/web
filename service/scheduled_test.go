// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestScheduled(t *testing.T) {
	a := assert.New(t)

	mgr := newManager(a, time.Local)
	a.Equal(0, len(mgr.Jobs()))

	mgr.AddAt("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(mgr.Jobs()))
}
