// SPDX-License-Identifier: MIT

package jsonp

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)

	j := JSONP("callback", 1)
	data, err := Marshal(j)
	a.NotError(err).Equal(string(data), "callback(1)")

	j = JSONP("", 1)
	data, err = Marshal(j)
	a.NotError(err).Equal(string(data), "1")

	data, err = Marshal(1)
	a.NotError(err).Equal(string(data), "1")
}
