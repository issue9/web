// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import (
	"testing"

	"github.com/issue9/assert"
)

func TestState_String(t *testing.T) {
	a := assert.New(t)

	a.Equal("faild", StateFaild.String())
	a.Equal("stop", StateStop.String())
	a.Equal("running", StateRunning.String())
	a.Equal("<unknown>", State(-10).String())
}
