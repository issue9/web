// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"testing"

	"github.com/issue9/assert"
)

func TestDefaultConfig(t *testing.T) {
	a := assert.New(t)

	conf := DefaultConfig()
	a.NotError(conf.Sanitize())
}

func TestConfig_Sanitize(t *testing.T) {
	a := assert.New(t)

	conf := &Config{EnvelopeState: EnvelopeStateEnable}
	a.Error(conf.Sanitize())
}
