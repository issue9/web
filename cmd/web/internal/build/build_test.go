// SPDX-License-Identifier: MIT

package build

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestReplaceVar(t *testing.T) {
	a := assert.New(t, false)

	args := []string{"build", "-o", "out.exe", "-ldflags", "-X=xxx", "./src"}
	replaceVar(args, "1.0.0", "f-c", "c")
	a.Equal(args, []string{"build", "-o", "out.exe", "-ldflags", "-X=xxx", "./src"})

	args = []string{"build", "-o", "out.exe", "-ldflags", "-X={{version}}", "./src"}
	replaceVar(args, "1.0.0", "f-c", "c")
	a.Equal(args, []string{"build", "-o", "out.exe", "-ldflags", "-X=1.0.0", "./src"})
}
