// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package create

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// 构建一个简单的终端交互界面
type asker struct {
	reader *bufio.Reader
	output io.Writer

	err error
}

// 声明 asker 变量
func newAsker(input io.Reader, output io.Writer) *asker {
	return &asker{
		reader: bufio.NewReader(input),
		output: output,
	}
}

func (a *asker) println(v ...interface{}) {
	if a.err != nil {
		return
	}

	_, a.err = fmt.Fprintln(a.output, v...)
}

func (a *asker) print(v ...interface{}) {
	if a.err != nil {
		return
	}

	_, a.err = fmt.Fprint(a.output, v...)
}

func (a *asker) printf(format string, v ...interface{}) {
	if a.err != nil {
		return
	}

	_, a.err = fmt.Fprintf(a.output, format, v...)
}

func (a *asker) read() string {
	v, err := a.reader.ReadString('\n')
	if err != nil {
		a.err = err
		return ""
	}

	return v[:len(v)-1]
}

// Ask 输出问题，并获取用户的回答内容
//
// q 显示的问题内容；
// def 表示默认值。
func (a *asker) Ask(q, def string) (string, error) {
	a.print(q)

	if def != "" {
		a.print("(", def, ")")
	}
	a.print(":")

	v := a.read()

	if a.err != nil {
		return "", a.err
	}

	if v == "" {
		v = def
	}
	return v, nil
}

// AskBool 输出二选一问题，并获取用户的回答内容
func (a *asker) AskBool(q string, def bool) (bool, error) {
	str := "Y"
	if !def {
		str = "N"
	}
	a.printf("%s(%s)\n", q, str)

	val := a.read()

	if a.err != nil {
		return false, a.err
	}

	switch strings.ToLower(val) {
	case "yes", "y":
		return true, nil
	case "no", "n":
		return false, nil
	default:
		return def, nil
	}
}

// AskSlice 输出一个单选问题，并获取用户的选择项
//
// q 表示问题内容；
// slice 表示可选的问题列表；
// def 表示默认项的索引，必须在 slice 之内。
func (a *asker) AskSlice(q string, slice []string, def int) (index int, err error) {
	a.println(q)
	for i, v := range slice {
		a.printf("(%d) %s\n", i, v)
	}
	a.print("请输入你的选择项:")

	val := a.read()

	if a.err != nil {
		return 0, a.err
	}

	if val == "" {
		return def, nil
	}

	return strconv.Atoi(val)
}

// AskMap 输出一个单选问题，并获取用户的选择项
//
// q 表示问题内容；
// maps 表示可选的问题列表；
// def 表示默认项的索引，必须在 maps 之内。
func (a *asker) AskMap(q string, maps map[string]string, def string) (key string, err error) {
	a.println(q)
	for k, v := range maps {
		a.printf("(%s) %s", k, v)
	}
	a.print("请输入你的选择项:")

	val := a.read()

	if a.err != nil {
		return "", a.err
	}

	if val == "" {
		val = def
	}
	return val, nil
}
