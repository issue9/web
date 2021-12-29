// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v5/group"

	"github.com/issue9/web/serialization/text"
)

func TestAccept(t *testing.T) {
	a := assert.New(t, false)

	srv := newServer(a, nil)
	err := srv.Mimetypes().Add(json.Marshal, json.Unmarshal, "text/json")
	a.NotError(err)

	ct := AcceptFilter(text.Mimetype, "text/json")
	r1 := srv.NewRouter("r1", "https://example.com", group.MatcherFunc(group.Any), ct)
	a.NotNil(r1)
	r1.Get("/path", func(*Context) Responser {
		return Status(http.StatusCreated)
	})

	s := rest.NewServer(a, srv.group, nil)
	s.Get("/path").
		Header("Accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated)

	s.Get("/path").
		Header("Accept", "application/json").
		Do(nil).
		Status(http.StatusNotAcceptable)

	s.Get("/path").
		Header("Accept", "text/json").
		Do(nil).
		Status(http.StatusCreated)
}
