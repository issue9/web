// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package selector

import (
	"strconv"
	"testing"
)

func BenchmarkRandom(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(false, 1), buildPeers(1))
	})

	b.Run("5", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(false, 1), buildPeers(5))
	})

	b.Run("10", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(false, 1), buildPeers(10))
	})
}

func BenchmarkWeightedRandom(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(true, 1), buildWeightedPeers(1))
	})

	b.Run("5", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(true, 1), buildWeightedPeers(5))
	})

	b.Run("10", func(b *testing.B) {
		benchSelector_Next(b, NewRandom(true, 1), buildWeightedPeers(10))
	})
}

func BenchmarkRoundRobin(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(false, 1), buildPeers(1))
	})

	b.Run("5", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(false, 1), buildPeers(5))
	})

	b.Run("10", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(false, 1), buildPeers(10))
	})
}

func BenchmarkWeightedRoundRobin(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(true, 1), buildWeightedPeers(1))
	})

	b.Run("5", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(true, 1), buildWeightedPeers(5))
	})

	b.Run("10", func(b *testing.B) {
		benchSelector_Next(b, NewRoundRobin(true, 1), buildWeightedPeers(10))
	})
}

func benchSelector_Next(b *testing.B, s Updateable, peers []Peer) {
	s.Update(peers...)
	for range b.N {
		if _, err := s.Next(); err != nil {
			b.Fatal(err)
		}
	}
}

func buildPeers(size int) []Peer {
	peers := make([]Peer, 0, size)
	for i := 0; i < size; i++ {
		peers = append(peers, NewPeer("http://localhost:808"+strconv.Itoa(i)))
	}
	return peers
}

func buildWeightedPeers(size int) []Peer {
	peers := make([]Peer, 0, size)
	for i := 0; i < size; i++ {
		peers = append(peers, NewWeightedPeer("http://localhost:808"+strconv.Itoa(i), i))
	}
	return peers
}
