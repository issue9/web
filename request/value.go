// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package request

import (
	"errors"
	"strconv"
)

type value interface {
	set(string) error
}

type intValue int

func (v *intValue) set(str string) error {
	i, err := strconv.Atoi(str)
	*v = intValue(i)
	return err
}

type int64Value int64

func (v *int64Value) set(str string) error {
	i, err := strconv.ParseInt(str, 10, 64)
	*v = int64Value(i)
	return err
}

// 大于 0 的 int64
type idValue int64

func (v *idValue) set(str string) error {
	i, err := strconv.ParseInt(str, 10, 64)
	*v = idValue(i)

	if err != nil {
		return err
	}

	if *v <= 0 {
		return errors.New("必须大于0")
	}

	return nil
}

type float64Value float64

func (v *float64Value) set(str string) error {
	i, err := strconv.ParseFloat(str, 64)
	*v = float64Value(i)
	return err
}

type stringValue string

func (v *stringValue) set(str string) error {
	*v = stringValue(str)
	return nil
}

type boolValue bool

func (v *boolValue) set(str string) error {
	i, err := strconv.ParseBool(str)
	*v = boolValue(i)
	return err
}
