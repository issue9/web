// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t)

	o := &Options{}
	a.NotError(o.sanitize())
	a.Equal(o.Location, time.Local).
		NotNil(o.groups).
		NotNil(o.Logs).
		NotNil(o.ResultBuilder).
		NotNil(o.Recovery)
}
