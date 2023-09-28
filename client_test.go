// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/compress"
	"github.com/issue9/web/servertest"
)

func TestClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, nil)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.NewRouter("default", nil).Get("/get", func(ctx *Context) Responser {
		return OK(&object{Name: "name"})
	})

	mts := []*Mimetype{
		{Type: "application/json", Marshal: marshalJSON, Unmarshal: json.Unmarshal, ProblemType: "application/problem+json"},
	}
	cps := []*Compress{
		{Name: "gzip", Compress: compress.NewGzipCompress(3), Types: []string{"application/*"}},
	}
	c, err := NewClient("http://localhost:8080", "application/json", json.Marshal, mts, cps)
	a.NotError(err).NotNil(c)

	resp := &object{}
	p := &RFC7807{}
	a.NotError(c.Get("/get", resp, p))
	a.Zero(p).Equal(resp, &object{Name: "name"})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Delete("/get", resp, p))
	a.Zero(resp).Equal(p.Type, ProblemMethodNotAllowed)

}
