// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMessage(t *testing.T) {
	a := assert.New(t)

	a.Equal(Message(-1000), CodeNotExists)
}
