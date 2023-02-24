// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

const tickTimer = 500 * time.Microsecond

// start 表示服务协程运行成功；
// p 用于触发 panic；
// err 用于触发 error；
func buildService() (f ServiceFunc, start, p, err chan struct{}) {
	start = make(chan struct{}, 1)
	p = make(chan struct{}, 1)
	err = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		inited := false
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel service")
				return ctx.Err()
			case <-p:
				fmt.Println("panic service")
				panic("service panic")
			case <-err:
				fmt.Println("error service")
				return errors.New("service error")
			default:
				fmt.Println("service at ", now)
				if !inited {
					inited = true
					start <- struct{}{}
				}
			}
		}
		return nil
	}, start, p, err
}

func TestServer_service(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	srv := s.Services()

	a.Equal(1, len(srv.Services())) // scheduled
	s1, start1, _, _ := buildService()
	srv.Add(localeutil.Phrase("srv1"), s1)
	a.Equal(2, len(srv.Services()))
	sched := srv.Services()[0]
	srv1 := srv.Services()[1]
	<-start1
	a.Equal(srv1.service, s1) // 并不会改变状态
	a.Equal(srv1.State(), Running).
		Equal(sched.State(), Running)

	// 运行中添加
	s2, start2, _, _ := buildService()
	srv.Add(localeutil.Phrase("srv2"), s2)
	a.Equal(3, len(srv.Services()))
	srv2 := srv.Services()[2]
	<-start2
	a.Equal(Running, srv2.State()) // 运行中添加自动运行服务

	s.Close(0)
	time.Sleep(500 * time.Millisecond) // 等待主服务设置状态值
	a.Equal(srv1.State(), Stopped)
	a.Equal(sched.State(), Stopped)
	a.Equal(srv2.State(), Stopped)
}

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	srv := s.Services()
	a.Equal(0, len(srv.Jobs()))

	srv.AddAt("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(srv.Jobs()))
}

func TestService_state(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := assert.New(t, false)
		s := newServer(a, nil)

		srv1, start, _, _ := buildService()
		s.Services().Add(localeutil.Phrase("srv1"), srv1)
		<-start
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		s1 := s.services[1]
		a.Equal(s1.State(), Running)

		s.Close(0)
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		a.Equal(s1.State(), Stopped)
	})

	t.Run("panic", func(t *testing.T) {
		a := assert.New(t, false)
		s := newServer(a, nil)

		srv1, start, p, _ := buildService()
		s.Services().Add(localeutil.Phrase("srv1"), srv1)
		<-start
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		s1 := s.services[1]
		a.Equal(s1.State(), Running)

		p <- struct{}{}
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		a.Equal(s1.State(), Failed).
			Contains(s1.Err().Error(), "service panic")

		s.Close(0)
	})

	t.Run("error", func(t *testing.T) {
		a := assert.New(t, false)
		s := newServer(a, nil)

		srv1, start, _, err := buildService()
		s.Services().Add(localeutil.Phrase("srv1"), srv1)
		<-start
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		s1 := s.services[1]
		a.Equal(s1.State(), Running)

		err <- struct{}{}
		time.Sleep(2 * time.Millisecond) // 等待主服务设置状态值
		a.Equal(s1.State(), Failed).
			Contains(s1.Err().Error(), "service error")

		s.Close(0)
	})
}
