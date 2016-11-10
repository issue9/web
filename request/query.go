// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import "net/http"

type Query struct {
	abortOnError bool
	errors       map[string]string
	values       map[string]value
	request      *http.Request
}

func NewQuery(r *http.Request, abortOnError bool) (*Query, error) {
	return &Query{
		abortOnError: abortOnError,
		errors:       map[string]string{},
		values:       make(map[string]value, 5),
		request:      r,
	}, nil
}

func (q *Query) parseOne(key string, val value) (ok bool) {
	v := q.request.FormValue(key)

	if len(v) == 0 { // 不存在，使用原来的值
		return true
	}

	if err := val.set(v); err != nil {
		q.errors[key] = err.Error()
		return false
	}
	return true
}

func (q *Query) Int(key string, def int) *int {
	i := new(int)
	q.IntVar(i, key, def)
	return i
}

func (q *Query) IntVar(i *int, key string, def int) {
	*i = def
	q.values[key] = (*intValue)(i)
}

func (q *Query) Int64(key string, def int64) *int64 {
	i := new(int64)
	q.Int64Var(i, key, def)
	return i
}

func (q *Query) Int64Var(i *int64, key string, def int64) {
	*i = def
	q.values[key] = (*int64Value)(i)
}

func (q *Query) Parse() map[string]string {
	for k, v := range q.values {
		ok := q.parseOne(k, v)
		if !ok && q.abortOnError {
			break
		}
	}

	return q.errors
}
