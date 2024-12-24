// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"
)

// start 表示服务协程运行成功；
// p 用于触发 panic；
// err 用于触发 error；
func buildService() (f ServiceFunc, start, p, err chan struct{}) {
	const tickTimer = 500 * time.Microsecond

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
	s := newTestServer(a)
	srv := s.Services()
	a.Equal(2, len(srv.services)) // scheduled, unique id generate

	s1, start1, _, _ := buildService()
	srv.Add(Phrase("srv1"), s1)
	a.Equal(3, len(srv.services))
	sched := srv.services[0]
	srv1 := srv.services[2]
	<-start1
	a.Equal(srv1.service, s1). // 并不会改变状态
					Equal(srv1.state, Running).
					Equal(sched.state, Running)

	s2, start2, _, _ := buildService()
	srv.Add(Phrase("srv2"), s2)
	a.Equal(4, len(srv.services))
	srv2 := srv.services[3]
	<-start2
	a.Equal(Running, srv2.state) // 运行中添加自动运行服务

	s3, start3, _, _ := buildService()
	del := srv.Add(Phrase("srv3"), s3)
	a.Equal(5, len(srv.services))
	srv3 := srv.services[4]
	<-start3
	a.Equal(Running, srv3.state)
	del()                          // 删除 s3
	a.Equal(4, len(srv.services)). // 确定删除
					Equal(Running, srv2.state). // 确定不会改变其它服务的状态
					Equal(srv3.state, Stopped)

	s.Close(0)
	time.Sleep(1000 * time.Millisecond) // 等待主服务设置状态值
	a.Equal(srv1.state, Stopped).
		Equal(sched.state, Stopped).
		Equal(srv2.state, Stopped)
}

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	srv := s.Services()
	a.Equal(0, len(srv.scheduled.Jobs()))

	srv.AddAt(Phrase("lang"), func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(srv.scheduled.Jobs()))

	// 查找翻译项是否正确
	var found bool
	srv.VisitJobs(func(j *Job) {
		p := s.Locale().NewPrinter(language.SimplifiedChinese)
		if !found {
			found = j.Title().LocaleString(p) == "cn"
		}
	})
	a.True(found)
}

func TestService_state(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		a := assert.New(t, false)
		s := newTestServer(a)

		srv1, start, _, _ := buildService()
		s.Services().Add(Phrase("srv1"), srv1)
		<-start
		s1 := s.Services().services[1]
		a.Equal(s1.state, Running)

		s.Close(0)
		a.Wait(500*time.Millisecond). // 等待主服务设置状态值
						Equal(s1.state, Stopped)
	})

	t.Run("panic", func(t *testing.T) {
		a := assert.New(t, false)
		s := newTestServer(a)

		srv1, start, p, _ := buildService()
		s.Services().Add(Phrase("srv1"), srv1)
		<-start
		s1 := s.Services().services[2]
		a.Equal(s1.state, Running)

		p <- struct{}{}
		a.Wait(200*time.Millisecond). // 等待主服务设置状态值
						Equal(s1.state, Failed).
						Contains(s1.err.Error(), "service panic")

		s.Close(0)
	})

	t.Run("error", func(t *testing.T) {
		a := assert.New(t, false)
		s := newTestServer(a)

		srv1, start, _, err := buildService()
		s.Services().Add(Phrase("srv1"), srv1)
		<-start
		s1 := s.Services().services[2]
		a.Equal(s1.state, Running)

		err <- struct{}{}
		a.Wait(200*time.Millisecond). // 等待主服务设置状态值
						Equal(s1.state, Failed).
						Contains(s1.err.Error(), "service error")

		s.Close(0)
	})
}
