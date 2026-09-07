// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert/v5"
)

func TestHTTP(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		NewError(http.StatusBadRequest, nil)
	})

	err10 := &cerr{"err10"}
	err11 := &cerr{"err11"}
	err := NewError(http.StatusBadRequest, errors.Join(err10, err11))
	a.NotNil(err).
		ErrorIs(err, err10).
		ErrorIs(err, err11)

	target1, ok := errors.AsType[*HTTP](err)
	a.True(ok).
		Equal(target1.Error(), errors.Join(err10, err11).Error()).
		Equal(target1.Status, http.StatusBadRequest)

	// 二次包装

	err2 := NewError(http.StatusBadGateway, err)
	a.ErrorIs(err2, err10)

	target2, ok := errors.AsType[*HTTP](err2)
	a.True(ok).
		ErrorIs(err2, err10).
		ErrorIs(err2, err11).
		Equal(target2.Status, http.StatusBadGateway)

	// 相同的状态码，返回原来的值。

	err3 := NewError(http.StatusBadRequest, err)
	a.ErrorIs(err3, err)
}
