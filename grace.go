// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"os"
	"os/signal"

	"github.com/issue9/logs"
)

var graced bool

// Grace 指定触发 Shutdown() 的信号，若为空，则任意信号都触发。
//
// NOTE: 传递空值，与不调用，其结果是不同的。
// 若是不调用，则不会处理任何信号；若是传递空值调用，则是处理任何要信号。
func Grace(sig ...os.Signal) {
	if graced {
		return
	}

	go func() {
		graced = true

		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)

		if err := Shutdown(); err != nil {
			logs.Error(err)
		}
	}()
}
