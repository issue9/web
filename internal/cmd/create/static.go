// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

var logs = []byte(`<?xml version="1.0" encoding="utf-8"?>
<logs>
    <!-- info内容，先缓存到一定10条，再一次性输出 -->
    <info prefix="INFO" flag="">
        <buffer size="100">
            <rotate filename="info-%Y%m%d.%i.log" dir="./logs/" size="5M" />
        </buffer>
    </info>

    <!-- debug日志 -->
    <debug>
        <buffer size="5">
            <rotate filename="debug-%Y%m%d.%i.log" dir="./logs/debug/" size="5M" />
        </buffer>
    </debug>

    <trace>
        <buffer size="5">
            <rotate filename="trace-%Y%m%d.%i.log" dir="./logs/trace/" size="5M" />
        </buffer>
    </trace>

    <warn>
        <rotate filename="warn-%Y%m%d.%i.log"  dir="./logs/warn/" size="5M" />
    </warn>

    <error>
        <rotate filename="error-%Y%m%d.%i.log"  dir="./logs/error/" size="5M" />
    </error>

    <critical>
        <rotate filename="critical-%Y%m%d.%i.log"  dir="./logs/critical/" size="5M" />
    </critical>
</logs>
`)

var maingo = []byte(`// 内容由 web 自动生成，可根据需求自由修改！

package main

const appconfig = "./appconfig"

import (
    "encoding/json"
    "encoding/xml"

    "github.com/issue9/web"

    "%s"
)

func main() {
    web.Init(appconfig)

    web.Mimetypes().AddMarshals(map[string]encoding.MarshaleFunc {
        "application/json": json.Marshal,
        "application/xml": xml.Marshal,
    })

    web.Mimetypes().AddUnmarshals(map[string]encoding.UnmarshaleFunc {
        "application/json": json.Unmarshal,
        "application/xml": xml.Unmarshal,
    })

    // 所有的模块初始化在此函数
    modules.Init()

    web.Fatal(2, web.Serve())
}
`)

var modulesgo = []byte(`// 内容由 web 自动生成，可根据需求自由修改！

package modules

// Init 所有模块的初始化操作可在此处进行。
func Init() {
    // TODO
}
`)
