// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import "context"

// TaskFunc 服务实际需要执行的函数
//
// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
// now 表示调用此函数的时间。
type TaskFunc func(ctx context.Context) error

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
