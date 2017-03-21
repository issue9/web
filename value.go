// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import "strconv"

type value interface {
	set(string) error
}

type intValue int

func (v *intValue) set(str string) error {
	i, err := strconv.Atoi(str)
	if err != nil { // 若出错，则不改变原来的值
		return err
	}

	*v = intValue(i)
	return nil
}

type int64Value int64

func (v *int64Value) set(str string) error {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil { // 若出错，则不改变原来的值
		return err
	}

	*v = int64Value(i)
	return nil
}

type float64Value float64

func (v *float64Value) set(str string) error {
	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*v = float64Value(i)
	return nil
}

type stringValue string

func (v *stringValue) set(str string) error {
	*v = stringValue(str)
	return nil
}

type boolValue bool

func (v *boolValue) set(str string) error {
	i, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}

	*v = boolValue(i)
	return nil
}
