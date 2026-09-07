// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package bufpool

import (
	"testing"

	"github.com/issue9/assert/v5"
)

func TestBufpool(t *testing.T) {
	a := assert.New(t, false)

	b := New()
	a.NotNil(b).Equal(b.Len(), 0)
	Put(b)
}
