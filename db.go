// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"github.com/issue9/orm"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/forward"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var dbs = map[string]*orm.DB{}

// 初始化数据配置
func initDB() {
	for k, v := range cfg.DB {
		if len(v.DSN) == 0 {
			panic("initDB:未指定dsn参数")
		}

		var d forward.Dialect
		switch v.Driver {
		case "sqlite3":
			d = dialect.Sqlite3()
		case "postgres":
			d = dialect.Postgres()
		case "mysql":
			d = dialect.Mysql()
		default:
			panic("initDB:未知道的db.Driver值:" + v.Driver)
		}

		db, err := orm.NewDB(v.Driver, v.DSN, k, d)
		if err != nil {
			panic(err)
		}

		dbs[k] = db
	}
}

// 返回一个orm.DB实例，若不存在，第二个参数返回false。
func GetDB(name string) (*orm.DB, bool) {
	db, found := dbs[name]
	return db, found
}
