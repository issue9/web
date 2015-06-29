// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"testing"

	"github.com/issue9/assert"
)

func TestInitDB(t *testing.T) {
	a := assert.New(t)

	// 未知道的Driver
	cfg.DB = map[string]*dbConfig{
		"db1": &dbConfig{Driver: "unknown", Prefix: "", DSN: "./testdata/db1.db"},
	}
	a.Panic(func() { initDB() })
	a.Equal(0, len(dbs))

	// 未指定DSN值
	cfg.DB = map[string]*dbConfig{
		"db1": &dbConfig{Driver: "sqlite3", Prefix: "", DSN: ""},
	}
	a.Panic(func() { initDB() })
	a.Equal(0, len(dbs))
}

func TestDB(t *testing.T) {
	a := assert.New(t)

	cfg.DB = map[string]*dbConfig{
		"db1": &dbConfig{Driver: "sqlite3", Prefix: "", DSN: "./testdata/db1.db"},
		"db2": &dbConfig{Driver: "sqlite3", Prefix: "prefix_", DSN: "./testdata/db2.db"},
	}

	a.NotPanic(func() { initDB() })
	a.Equal(2, len(dbs))

	db := DB("db1")
	a.Equal(db, dbs["db1"])

	db = DB("db2")
	a.Equal(db, dbs["db2"])

	db = DB("nil")
	a.Nil(db)
}
