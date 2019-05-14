// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 简单的辅助功能命令行工具。
package main

import (
	"os"

	"github.com/issue9/web/internal/cmd/build"
	"github.com/issue9/web/internal/cmd/command"
	"github.com/issue9/web/internal/cmd/create"
	"github.com/issue9/web/internal/cmd/release"
	"github.com/issue9/web/internal/cmd/version"
	"github.com/issue9/web/internal/cmd/watch"
)

// 帮助信息的输出通道
var output = os.Stdout

func main() {
	command.Register("version", version.Do, version.Usage)
	command.Register("watch", watch.Do, watch.Usage)
	command.Register("create", create.Do, create.Usage)
	command.Register("build", build.Do, build.Usage)
	command.Register("release", release.Do, release.Usage)

	if err := command.Exec(output); err != nil {
		panic(err)
	}
}
