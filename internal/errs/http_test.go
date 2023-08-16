// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestHTTP(t *testing.T) {
	a := assert.New(t, false)

	err1 := &cerr{"abc"}
	err := NewHTTPError(http.StatusBadRequest, err1)
	a.NotNil(err).
		ErrorIs(err, err1)

	var target1 *HTTP
	a.True(errors.As(err, &target1)).
		Equal(target1.Error(), err1.Error()).
		Equal(target1.Status, http.StatusBadRequest)

	// 二次包装

	err = NewHTTPError(http.StatusBadGateway, err)
	a.ErrorIs(err, err1)

	var target2 *HTTP
	a.True(errors.As(err, &target2)).
		Equal(target2.Error(), err1.Error()).
		Equal(target2.Status, http.StatusBadGateway)
}
