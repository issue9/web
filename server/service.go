// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/scheduled"

	"github.com/issue9/web"
)

type (
	services struct {
		s         *httpServer
		services  []*service
		ctx       context.Context
		scheduled *scheduled.Server
		jobTitles map[string]web.LocaleStringer
	}

	service struct {
		s       *services
		title   web.LocaleStringer
		service web.Service
		err     error // 保存上次的出错内容，不会清空该值。

		state    web.State
		stateMux sync.Mutex
	}
)

func (srv *httpServer) initServices() {
	ctx, cancel := context.WithCancelCause(context.Background())
	srv.OnClose(func() error { cancel(http.ErrServerClosed); return nil })

	srv.services = &services{
		s:         srv,
		services:  make([]*service, 0, 5),
		ctx:       ctx,
		scheduled: scheduled.NewServer(srv.Location(), srv.logs.ERROR(), srv.logs.DEBUG()),
		jobTitles: make(map[string]web.LocaleStringer, 10),
	}
	srv.services.Add(web.StringPhrase("scheduler jobs"), srv.services.scheduled)
}

func (srv *service) setState(s web.State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

func (srv *service) goServe() {
	if srv.state != web.Running {
		srv.setState(web.Running)
		go srv.serve()
	}
}

func (srv *service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.s.Logs().ERROR().Error(srv.err)
			srv.setState(web.Failed)
		}
	}()
	srv.err = srv.service.Serve(srv.s.ctx)
	state := web.Stopped
	if !errors.Is(srv.err, context.Canceled) {
		srv.s.s.Logs().ERROR().Error(srv.err)
		state = web.Failed
	}

	srv.setState(state)
}

func (srv *httpServer) Services() web.Services { return srv.services }

func (srv *services) Add(title web.LocaleStringer, f web.Service) {
	s := &service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.services = append(srv.services, s)
	s.goServe()
}

func (srv *services) AddFunc(title web.LocaleStringer, f func(context.Context) error) {
	srv.Add(title, web.ServiceFunc(f))
}

func (srv *services) Visit(visit func(title web.LocaleStringer, state web.State, err error)) {
	for _, s := range srv.services {
		visit(s.title, s.state, s.err)
	}
}

func (srv *services) AddCron(title web.LocaleStringer, f web.JobFunc, spec string, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.Cron(id, f, spec, delay)
}

func (srv *services) AddTicker(title web.LocaleStringer, job web.JobFunc, dur time.Duration, imm, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.Tick(id, job, dur, imm, delay)
}

func (srv *services) AddAt(title web.LocaleStringer, job web.JobFunc, at time.Time, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.At(id, job, at, delay)
}

func (srv *services) AddJob(title web.LocaleStringer, job web.JobFunc, scheduler web.Scheduler, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.New(id, job, scheduler, delay)
}

func (srv *services) VisitJobs(visit func(web.LocaleStringer, time.Time, time.Time, web.State, bool, error)) {
	for _, j := range srv.scheduled.Jobs() {
		visit(srv.jobTitles[j.Name()], j.Prev(), j.Next(), j.State(), j.Delay(), j.Err())
	}
}
