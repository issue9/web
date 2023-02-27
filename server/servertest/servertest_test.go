// SPDX-License-Identifier: MIT

package servertest_test

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/server/servertest"
)

type server struct {
	exit chan struct{}
}

func (s *server) Serve() error {
	<-s.exit
	return http.ErrServerClosed
}

func (s *server) Close() error {
	s.exit <- struct{}{}
	return nil
}

func TestRun(t *testing.T) {
	a := assert.New(t, false)
	s := &server{exit: make(chan struct{}, 1)}

	wait := servertest.Run(a, s)
	t.Log("before close")
	s.Close()
	t.Log("after close")
	wait()
}
