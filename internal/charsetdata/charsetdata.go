// SPDX-License-Identifier: MIT

// Package charsetdata 用于测试的字符集数据
package charsetdata

import (
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	GBKString1 = "中文1,11"
	GBKString2 = "中文2,22"
)

var (
	GBKData1, GBKData2 []byte
)

func init() {
	reader := transform.NewReader(strings.NewReader(GBKString1), simplifiedchinese.GBK.NewEncoder())
	gbkData, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	GBKData1 = gbkData

	reader = transform.NewReader(strings.NewReader(GBKString2), simplifiedchinese.GBK.NewEncoder())
	gbkData, err = ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	GBKData2 = gbkData
}
