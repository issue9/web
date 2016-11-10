// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"testing"

	"github.com/issue9/assert"
)

func TestParam_Int(t *testing.T) {
	a := assert.New(t)

	params := map[string]string{
		"p1": "5",
	}
	p := &Param{
		abortOnError: false,
		errors:       map[string]string{},
		values:       make(map[string]value, len(params)),
		params:       params,
	}

	p1 := p.Int("p1")
	a.Equal(len(p.Parse()), 0).Equal(*p1, 5)
}
