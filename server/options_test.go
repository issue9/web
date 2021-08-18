// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t)

	var o *Options
	oo, err := o.sanitize()
	a.NotError(err).NotNil(oo)

	o = &Options{}
	oo, err = o.sanitize()
	a.NotError(err).NotNil(oo)
	a.Equal(oo.Location, time.Local)
	a.NotNil(oo.groups)
	a.NotNil(oo.Logs)
}
