// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package dev

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestDev(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(Filename("web.yaml"), "web_development.yaml").
		Equal(Filename("web"), "web_development")
}
