// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

func clearMesages() {
	messages = map[int]message{}
}

func TestGetStatus(t *testing.T) {
	a := assert.New(t)

	a.Equal(getStatus(100), 100)
	a.Equal(getStatus(200), 200)
	a.Equal(getStatus(211), 211)
	a.Equal(getStatus(9011), 901)
	a.Equal(getStatus(9099), 909)
}

func TestNewMessage(t *testing.T) {
	a := assert.New(t)

	a.Error(NewMessage(99, "99"))      // 必须大于等于 100
	a.NotError(NewMessage(100, "100")) // 必须大于等于 100

	clearMesages()
}
