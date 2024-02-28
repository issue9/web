// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package errs

import "errors"

type HTTP struct {
	Status  int
	Message error
}

func (err *HTTP) Error() string { return err.Message.Error() }

func NewError(status int, err error) error {
	if err == nil {
		panic("err 不能为空")
	}

	var herr *HTTP
	if errors.As(err, &herr) {
		if herr.Status == status {
			return herr
		}
		return &HTTP{Status: status, Message: herr.Message}
	}

	return &HTTP{Status: status, Message: err}
}

func (err *HTTP) Unwrap() error { return err.Message }
