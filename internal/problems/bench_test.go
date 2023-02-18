// SPDX-License-Identifier: MIT

package problems

import "testing"

func BenchmarkRFC7807Pool_New(b *testing.B) {
	pool := NewRFC7807Pool[*ctxDemo]()
	ctx := &ctxDemo{}

	for i := 0; i < b.N; i++ {
		p := pool.New("id", 200, "title", "detail")
		p.With("custom", "custom")
		p.AddParam("p1", "v1")
		p.Apply(ctx)
	}
}
