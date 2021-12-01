// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v2"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o := &Options{}
	a.NotError(o.sanitize())
	a.Equal(o.Location, time.Local).
		NotNil(o.group).
		NotNil(o.Logs).
		NotNil(o.ResultBuilder)
}
