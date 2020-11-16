// SPDX-License-Identifier: MIT

package web

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		New(logs.New(), nil)
	})

	web, err := New(logs.New(), &Config{})
	a.NotError(err).NotNil(web)
}
