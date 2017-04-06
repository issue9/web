// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package install 提供程序安装时的一些功能。
//  install.New("admin", install, "dep1", "dep2")
//  func install(log install.Logger) {
//      log.Println()
//      log.Info()
//  }
package install

import "github.com/issue9/web/modules"

var ms = modules.New()

type InstallFunc func(*Logger) error

func New(name string, fn InstallFunc, deps ...string) {
	f := func() error {
		l := &Logger{}

		l.Infof("===== 安装[%v]...\n")

		if err := fn(l); err != nil {
			l.Errorln("安装失败", err)
		} else {
			l.Success()
		}

		return nil
	}

	ms.New(name, f, deps...)
}

func Install() error {
	return ms.Init()
}
