// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/issue9/assert"
)

func TestQuery_Int(t *testing.T) {
	a := assert.New(t)

	form := url.Values(map[string][]string{
		"q1": []string{"1"},
	})
	q := &Query{
		abortOnError: false,
		errors:       map[string]string{},
		values:       make(map[string]value, len(form)),
		request:      &http.Request{Form: form},
	}

	q1 := q.Int("p1", 12)
	msgs := q.Parse()
	a.Equal(len(msgs), 0, msgs).Equal(*q1, 1)
}
