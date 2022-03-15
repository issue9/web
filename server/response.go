// SPDX-License-Identifier: MIT

package server

type (
	Responser interface {
		Apply(*Context) error
	}

	status int

	Object struct {
		status  int
		body    any
		headers map[string]string
	}
)

func Status(code int) Responser { return status(code) }

func (s status) Apply(ctx *Context) error {
	ctx.WriteHeader(int(s))
	return nil
}

func (o *Object) Apply(ctx *Context) error {
	return ctx.Marshal(o.status, o.body, o.headers)
}

func Body(status int, body any) *Object {
	return &Object{status: status, body: body}
}

func (o *Object) Header(k, v string) *Object {
	if o.headers == nil {
		o.headers = make(map[string]string, 3)
	}
	o.headers[k] = v
	return o
}
