// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

var logs = []byte(`<?xml version="1.0" encoding="utf-8" ?>
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
