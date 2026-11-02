// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
)

func TestCreateMapCollector(t *testing.T) {
	a := assert.New(t, false)

	m, c := CreateMapCollector()
	a.NotNil(m).
		NotNil(c).
		Empty(m)

	r1 := localeutil.Phrase("reason1")
	r2 := localeutil.Phrase("reason2")
	r3 := localeutil.Phrase("reason3")

	c("name1", r1)
	a.Equal(m, map[string][]localeutil.Stringer{"name1": []localeutil.Stringer{r1}})

	c("name1", r2)
	a.Equal(m, map[string][]localeutil.Stringer{"name1": []localeutil.Stringer{r1, r2}})

	c("name2", r2)
	a.Equal(m, map[string][]localeutil.Stringer{
		"name1": []localeutil.Stringer{r1, r2},
		"name2": []localeutil.Stringer{r2},
	})

	c("name2", r3)
	a.Equal(m, map[string][]localeutil.Stringer{
		"name1": []localeutil.Stringer{r1, r2},
		"name2": []localeutil.Stringer{r2, r3},
	})
}
