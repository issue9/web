// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package registry

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web/selector"
)

func BenchmarkMarshalPeers(b *testing.B) {
	a := assert.New(b, false)

	peers := []selector.Peer{
		selector.NewPeer("http://localhost:8080"),
		selector.NewPeer("http://localhost:8081"),
		selector.NewPeer("http://localhost:8082"),
		selector.NewPeer("http://localhost:8083"),
		selector.NewPeer("http://localhost:8084"),
	}

	b.Run("marshalPeers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := marshalPeers(peers); err != nil {
				b.Fatal(err)
			}
		}
	})

	data, err := marshalPeers(peers)
	a.NotError(err).NotEmpty(data)
	nb := func() selector.Peer { return selector.NewPeer("") }

	b.Run("unmarshalPeers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := unmarshalPeers(nb, data); err != nil {
				b.Fatal(err)
			}
		}
	})
}
