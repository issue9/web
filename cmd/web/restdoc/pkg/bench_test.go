// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func BenchmarkPackages_TypeOf_Int(b *testing.B) {
	a := assert.New(b, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.ScanDir(context.Background(), "./testdir", true)
	ctx := context.Background()
	for range b.N {
		p.TypeOf(ctx, "github.com/issue9/web/restdoc/pkg.Int")
	}
}

func BenchmarkPackages_TypeOf_S(b *testing.B) {
	a := assert.New(b, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.ScanDir(context.Background(), "./testdir", true)
	ctx := context.Background()
	for range b.N {
		p.TypeOf(ctx, "github.com/issue9/web/restdoc/pkg.S")
	}
}
