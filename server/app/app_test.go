// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/localeutil"
	"github.com/kardianos/service"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
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

func TestDaemonConfig_toServiceConfig(t *testing.T) {
	a := assert.New(t, false)

	d := &DaemonConfig{
		DisplayName: web.Phrase("displayName"),
		Description: web.Phrase("desc"),
		UserName:    "username",
	}

	p := localeutil.NewPrinter(catalog.NewBuilder(), language.SimplifiedChinese)
	s := d.toServiceConfig("id", p)
	a.NotNil(s).
		Equal(s.DisplayName, "displayName").
		Equal(s.Description, "desc").
		Equal(s.Name, "id")
}
