// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNewMessage(t *testing.T) {
	a := assert.New(t)

	err := NewMessage("test")
	a.Error(err)
	msg, ok := err.(message)
	a.True(ok).
		Equal(string(msg), "test")
}
