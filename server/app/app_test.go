// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/kardianos/service"
)

type empty struct{}

var _ service.Interface = &app{}

func TestStatusString(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(statusString(service.StatusRunning), "running").
		Equal(statusString(service.StatusStopped), "stopped").
		Equal(statusString(service.StatusUnknown), "unknown").
		Equal(statusString(10), "unknown")
}
