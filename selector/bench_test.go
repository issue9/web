// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func BenchmarkRandom(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1", func(b *testing.B) {
		s := NewRandom(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		benchSelector_Next(b, s)
	})

	b.Run("5", func(b *testing.B) {
		s := NewRandom(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		a.NotError(s.Add(NewPeer("http://localhost:8081")))
		a.NotError(s.Add(NewPeer("http://localhost:8082")))
		a.NotError(s.Add(NewPeer("http://localhost:8083")))
		a.NotError(s.Add(NewPeer("http://localhost:8084")))
		benchSelector_Next(b, s)
	})

	b.Run("10", func(b *testing.B) {
		s := NewRandom(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		a.NotError(s.Add(NewPeer("http://localhost:8081")))
		a.NotError(s.Add(NewPeer("http://localhost:8082")))
		a.NotError(s.Add(NewPeer("http://localhost:8083")))
		a.NotError(s.Add(NewPeer("http://localhost:8084")))
		a.NotError(s.Add(NewPeer("http://localhost:8085")))
		a.NotError(s.Add(NewPeer("http://localhost:8086")))
		a.NotError(s.Add(NewPeer("http://localhost:8087")))
		a.NotError(s.Add(NewPeer("http://localhost:8088")))
		a.NotError(s.Add(NewPeer("http://localhost:8089")))
		benchSelector_Next(b, s)
	})
}

func BenchmarkWeightedRandom(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1", func(b *testing.B) {
		s := NewRandom(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		benchSelector_Next(b, s)
	})

	b.Run("5", func(b *testing.B) {
		s := NewRandom(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		a.NotError(s.Add(NewPeer("http://localhost:8081", 3)))
		a.NotError(s.Add(NewPeer("http://localhost:8082", 5)))
		a.NotError(s.Add(NewPeer("http://localhost:8083", 8)))
		a.NotError(s.Add(NewPeer("http://localhost:8084", 9)))
		benchSelector_Next(b, s)
	})

	b.Run("10", func(b *testing.B) {
		s := NewRandom(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		a.NotError(s.Add(NewPeer("http://localhost:8081", 3)))
		a.NotError(s.Add(NewPeer("http://localhost:8082", 5)))
		a.NotError(s.Add(NewPeer("http://localhost:8083", 9)))
		a.NotError(s.Add(NewPeer("http://localhost:8084", 11)))
		a.NotError(s.Add(NewPeer("http://localhost:8085", 22)))
		a.NotError(s.Add(NewPeer("http://localhost:8086", 33)))
		a.NotError(s.Add(NewPeer("http://localhost:8087", 44)))
		a.NotError(s.Add(NewPeer("http://localhost:8088", 55)))
		a.NotError(s.Add(NewPeer("http://localhost:8089", 66)))
		benchSelector_Next(b, s)
	})
}

func BenchmarkRoundRobin(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1", func(b *testing.B) {
		s := NewRoundRobin(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		benchSelector_Next(b, s)
	})

	b.Run("5", func(b *testing.B) {
		s := NewRoundRobin(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		a.NotError(s.Add(NewPeer("http://localhost:8081")))
		a.NotError(s.Add(NewPeer("http://localhost:8082")))
		a.NotError(s.Add(NewPeer("http://localhost:8083")))
		a.NotError(s.Add(NewPeer("http://localhost:8084")))
		benchSelector_Next(b, s)
	})

	b.Run("10", func(b *testing.B) {
		s := NewRoundRobin(false, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080")))
		a.NotError(s.Add(NewPeer("http://localhost:8081")))
		a.NotError(s.Add(NewPeer("http://localhost:8082")))
		a.NotError(s.Add(NewPeer("http://localhost:8083")))
		a.NotError(s.Add(NewPeer("http://localhost:8084")))
		a.NotError(s.Add(NewPeer("http://localhost:8085")))
		a.NotError(s.Add(NewPeer("http://localhost:8086")))
		a.NotError(s.Add(NewPeer("http://localhost:8087")))
		a.NotError(s.Add(NewPeer("http://localhost:8088")))
		a.NotError(s.Add(NewPeer("http://localhost:8089")))
		benchSelector_Next(b, s)
	})
}

func BenchmarkWeightedRoundRobin(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1", func(b *testing.B) {
		s := NewRoundRobin(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		benchSelector_Next(b, s)
	})

	b.Run("5", func(b *testing.B) {
		s := NewRoundRobin(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		a.NotError(s.Add(NewPeer("http://localhost:8081", 3)))
		a.NotError(s.Add(NewPeer("http://localhost:8082", 5)))
		a.NotError(s.Add(NewPeer("http://localhost:8083", 8)))
		a.NotError(s.Add(NewPeer("http://localhost:8084", 9)))
		benchSelector_Next(b, s)
	})

	b.Run("10", func(b *testing.B) {
		s := NewRoundRobin(true, 1)
		a.NotError(s.Add(NewPeer("http://localhost:8080", 1)))
		a.NotError(s.Add(NewPeer("http://localhost:8081", 3)))
		a.NotError(s.Add(NewPeer("http://localhost:8082", 5)))
		a.NotError(s.Add(NewPeer("http://localhost:8083", 9)))
		a.NotError(s.Add(NewPeer("http://localhost:8084", 11)))
		a.NotError(s.Add(NewPeer("http://localhost:8085", 22)))
		a.NotError(s.Add(NewPeer("http://localhost:8086", 33)))
		a.NotError(s.Add(NewPeer("http://localhost:8087", 44)))
		a.NotError(s.Add(NewPeer("http://localhost:8088", 55)))
		a.NotError(s.Add(NewPeer("http://localhost:8089", 66)))
		benchSelector_Next(b, s)
	})
}

func benchSelector_Next(b *testing.B, s Selector) {
	for i := 0; i < b.N; i++ {
		if _, err := s.Next(); err != nil {
			b.Fatal(err)
		}
	}
}
