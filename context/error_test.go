// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func TestTraceStack(t *testing.T) {
	a := assert.New(t)

	str := traceStack(1, "message", 12)
	a.True(strings.HasPrefix(str, "message12"))
	a.True(strings.Contains(str, "error_test.go")) // 肯定包含当前文件名
}