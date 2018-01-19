// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/issue9/web/encoding"
)

func TestAddMarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(v interface{}) ([]byte, error) { return nil, nil }

	a.Equal(1, len(marshals))

	a.NotError(AddMarshal("n1", f1))
	a.NotError(AddMarshal("n2", f1))
	a.Equal(AddMarshal("n2", f1), ErrExists)
	a.Equal(3, len(marshals))
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(data []byte, v interface{}) error { return nil }

	a.Equal(1, len(unmarshals))

	a.NotError(AddUnmarshal("n1", f1))
	a.NotError(AddUnmarshal("n2", f1))
	a.Equal(AddUnmarshal("n2", f1), ErrExists)
	a.Equal(3, len(unmarshals))
}

func TestAddCharset(t *testing.T) {
	a := assert.New(t)
	e := simplifiedchinese.GBK

	a.Equal(1, len(charset))
	a.Nil(charset[encoding.DefaultCharset])

	a.NotError(AddCharset("n1", e))
	a.NotError(AddCharset("n2", e))
	a.Equal(AddCharset("n2", e), ErrExists)
	a.Equal(3, len(charset))
}
