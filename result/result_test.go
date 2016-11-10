// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	r := New(-2)
	a.Equal(r.Message, codeNotExists)

	r = New(RegisterMessage(400, "400"))
	a.Equal(r.Message, "400")
}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	r := New(400 * scale)
	a.False(r.HasDetail())

	r.Add("field", "message")
	a.True(r.HasDetail())
}

func TestResult_IsError(t *testing.T) {
	a := assert.New(t)

	r := New(400 * scale)
	a.True(r.IsError())

	r = New(300 * scale)
	a.False(r.IsError())
}
