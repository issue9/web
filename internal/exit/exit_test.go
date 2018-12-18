// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package exit

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestContext(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() { Context(5) })
	a.Panic(func() { Context(0) })

	func() {
		defer func() {
			msg := recover()
			val, ok := msg.(HTTPStatus)
			a.True(ok).Equal(val, http.StatusNotFound)
		}()

		Context(http.StatusNotFound)
	}()
}
