// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package bufpool

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestBufpool(t *testing.T) {
	a := assert.New(t, false)

	b := New()
	a.NotNil(b).Equal(b.Len(), 0)
	Put(b)
}
