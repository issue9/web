// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"encoding/json/v2"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/goccy/go-yaml"
	"github.com/issue9/assert/v6"
)

type object struct {
	Name string `json:"name"`
	Age  int
}

func BenchmarkProblems_Get(b *testing.B) {
	a := assert.New(b, false)
	ps := New[render]()

	var item *Item
	for b.Loop() {
		item = ps.Get(BadGateway)
	}
	a.NotNil(item)
}

func BenchmarkNewProblem(b *testing.B) {
	ps := New[render]()
	b.ReportAllocs()
	for b.Loop() {
		p := ps.NewProblem()
		p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
		ps.Free(p)
	}
}

func BenchmarkProblem_JSON(b *testing.B) {
	benchProblem_marshal_unmarshal(b, func(a any) ([]byte, error) { return json.Marshal(a) }, func(b []byte, a any) error { return json.Unmarshal(b, a) })
}

func BenchmarkProblem_CBOR(b *testing.B) {
	benchProblem_marshal_unmarshal(b, cbor.Marshal, cbor.Unmarshal)
}

func BenchmarkProblem_YAML(b *testing.B) {
	benchProblem_marshal_unmarshal(b, yaml.Marshal, yaml.Unmarshal)
}

func benchProblem_marshal_unmarshal(b *testing.B, marshal func(any) ([]byte, error), unmarshal func([]byte, any) error) {
	a := assert.New(b, false)
	ps := New[render]()

	p := ps.NewProblem()
	p.Type = "https://iana.org/assignments/http-problem-types#date"
	p.Title = "http-problem-types#date"
	p.Detail = "https://iana.org/assignments/http-problem-types#date"
	p.Status = 400
	p.WithInstance("500")
	p.WithExtensions("extension")
	p.WithParam("p1", "p1-1")
	p.WithParam("p1", "p1-2")
	p.WithParam("p2", "p2")

	var data []byte
	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			data, _ = marshal(p)
		}
	})

	out := ps.NewProblem()
	b.Run("unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			unmarshal(data, out)
		}
	})

	a.Equal(p, out)
}
