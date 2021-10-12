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
	a.Equal(o.Location, time.Local)
	a.NotNil(o.groups)
	a.NotNil(o.Logs)
}
