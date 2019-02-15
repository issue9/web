// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"time"
)

// TaskFunc 服务实际需要执行的函数
//
// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
// now 表示调用此函数的时间。
type TaskFunc func(ctx context.Context, now time.Time) error

// NextFunc 获取下一次服务的执行时间。
type NextFunc func() <-chan time.Time

// State 服务的状态值
type State int8

// 几种可能的状态值
const (
	StateWating  State = iota + 1 // 等待下次运行，默认状态
	StateRunning                  // 正在运行
	StateStop                     // 正常停止，将不再执行后续操作
	StateFaild                    // 出错，不再执行后续操作
)

// ErrorHandling 出错时的处理方式
type ErrorHandling int8

// 定义几种出错时的处理方式
const (
	ContinueOnError ErrorHandling = iota + 1
	ExitOnError
)

// PrevHandling 表示在下一次执行时，如果前一任务未完成，如何处理。
type PrevHandling int8

// PrevHandling 的几种常量
const (
	AbortOnNext PrevHandling = iota + 1
	ContinueOnNext
)

func (s State) String() string {
	switch s {
	case StateWating:
		return "wating"
	case StateRunning:
		return "running"
	case StateStop:
		return "stop"
	case StateFaild:
		return "faild"
	}

	return "<unknown>"
}

// Tick 定时功能的 NextFunc。
//
// 是对 time.Tick 的简单封闭。
func Tick(d time.Duration) NextFunc {
	return NextFunc(func() <-chan time.Time {
		return time.Tick(d)
	})
}
