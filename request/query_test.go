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
		"q2": []string{"21", "22"},
		"q4": []string{"four"},
	})
	r := &http.Request{Form: form}

	q := &Query{
		abortOnError: false,
		errors:       map[string]string{},
		values:       make(map[string]value, len(form)),
		request:      r,
	}

	q1 := q.Int("q1", 12)
	msgs := q.Parse()
	a.Equal(len(msgs), 0).Equal(*q1, 1)

	q2 := q.Int64("q2", 12)
	msgs = q.Parse()
	a.Equal(len(msgs), 0).Equal(*q2, 21)

	q3 := q.Int64("q3", 32)
	msgs = q.Parse()
	a.Equal(len(msgs), 0).Equal(*q3, 32)

	// 出错的情况下，返回默认值
	q4 := q.Int64("q4", 32)
	msgs = q.Parse()
	a.Equal(len(msgs), 1).Equal(*q4, 32)
}
