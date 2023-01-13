// SPDX-License-Identifier: MIT

package cache

import (
	"strconv"
	"testing"

	"github.com/issue9/assert/v3"
)

type object int

func (o object) MarshalCache() ([]byte, error) {
	return []byte(strconv.Itoa(int(o))), nil
}

func (o *object) UnmarshalCache(bs []byte) error {
	n, err := strconv.Atoi(string(bs))
	if err != nil {
		return err
	}
	*o = object(n)
	return nil
}

func TestSerializer(t *testing.T) {
	a := assert.New(t, false)

	num := 5
	data, err := Marshal(num)
	a.NotError(err).NotNil(data)
	var num2 int
	a.NotError(Unmarshal(data, &num2))
	a.Equal(num2, num)

	o := object(6)
	data, err = Marshal(o)
	a.NotError(err).NotNil(data)
	var o2 object
	a.NotError(Unmarshal(data, &o2))
	a.Equal(o2, o)
}
