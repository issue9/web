// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
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
	s := newTestServer(a, false)
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

	s.close()                           // 并未调用 Serve，直接调用 Close 会阻塞，所以直接调用 close 结束。
	time.Sleep(1000 * time.Millisecond) // 等待主服务设置状态值
	a.Equal(srv1.state, Stopped).
		Equal(sched.state, Stopped).
		Equal(srv2.state, Stopped)
}

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)
	srv := s.Services()
	a.Equal(0, len(slices.Collect(srv.scheduled.Jobs())))

	srv.AddAt(Phrase("lang"), func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now())
	a.Equal(1, len(slices.Collect(srv.scheduled.Jobs())))

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
		ts := newTestServer(a, false)

		f, start, _, _ := buildService()
		ts.Services().Add(Phrase("normal service"), f)
		<-start
		ss := ts.Services().services
		s := ss[len(ss)-1]
		a.Equal(s.state, Running)

		ts.close()                         // 触发 context.WithCancel
		time.Sleep(200 * time.Millisecond) // 等待 context.Context 链路完成
		a.Equal(s.state, Stopped)
	})

	t.Run("panic", func(t *testing.T) {
		a := assert.New(t, false)
		ts := newTestServer(a, false)

		f, start, p, _ := buildService()
		ts.Services().Add(Phrase("panic service"), f)
		<-start
		ss := ts.Services().services
		s := ss[len(ss)-1]
		a.StrictEqual(s.state, Running)

		p <- struct{}{}
		time.Sleep(200 * time.Millisecond) // 等待 p 被取出且 s 的状态设置完成
		a.StrictEqual(s.state, Failed).Contains(s.err.Error(), "service panic")

		ts.close()
	})

	t.Run("error", func(t *testing.T) {
		a := assert.New(t, false)
		ts := newTestServer(a, false)

		f, start, _, err := buildService()
		ts.Services().Add(Phrase("error service"), f)
		<-start
		ss := ts.Services().services
		s := ss[len(ss)-1]
		a.StrictEqual(s.state, Running)

		err <- struct{}{}
		time.Sleep(200 * time.Millisecond) // 等待 err 被取出且 s 的状态设置完成
		a.StrictEqual(s.state, Failed).NotNil(s.err).Contains(s.err.Error(), "service error")

		ts.close()
	})
}
