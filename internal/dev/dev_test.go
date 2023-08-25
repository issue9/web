// SPDX-License-Identifier: MIT

package dev

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestDev(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(Filename("web.yaml"), "web_development.yaml").
		Equal(Filename("web"), "web_development")
}
