// SPDX-License-Identifier: MIT

package server

import (
	"fmt"

	"github.com/issue9/events"
)

// Subscriber 订阅者函数
//
// 每个订阅函数都是通过 go 异步执行。
//
// data 为事件传递过来的数据，可能存在多个订阅者，
// 用户不应该直接修改 data 数据，否则结果是未知的。
type Subscriber = events.Subscriber

type publisher struct {
	name string
	s    *Server
	p    events.Publisher
}

func (p *publisher) Publish(sync bool, data interface{}) error {
	return p.p.Publish(sync, data)
}

func (p *publisher) Destory() {
	delete(p.s.events, p.name)
	p.p.Destory()
}

// Publisher 创建事件发布者
func (srv *Server) Publisher(name string) events.Publisher {
	if _, found := srv.events[name]; found {
		panic(fmt.Sprintf("事件 %s 已经存在", name))
	}

	p, e := events.New()
	srv.events[name] = e
	return &publisher{
		name: name,
		s:    srv,
		p:    p,
	}
}

// Eventer 返回指定名称的事件处理对象
//
// name 表示事件名称，该名称必须在 Publisher 中创建。
func (srv *Server) Eventer(name string) events.Eventer { return srv.events[name] }

// AttachEvent 订阅指定事件
//
// 返回的值可用于取消订阅。
func (srv *Server) AttachEvent(name string, s Subscriber) (int, error) {
	return srv.Eventer(name).Attach(s)
}

// DetachEvent 取消对某事件的订阅
func (srv *Server) DetachEvent(name string, id int) { srv.Eventer(name).Detach(id) }
