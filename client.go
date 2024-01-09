// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"io"
	"net/http"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/status"
	"github.com/issue9/web/selector"
)

// Client 用于访问远程的客户端
//
// NOTE: 远程如果不是 [Server] 实现的服务，可能无法正确处理返回对象。
type Client struct {
	client   *http.Client
	codec    *Codec
	selector selector.Selector

	marshal     func(any) ([]byte, error)
	marshalName string

	requestIDKey string
	requestIDGen func() string
}

// NewClient 创建 Client 实例
//
// client 要以为空，表示采用 &http.Client{} 作为默认值；
// marshalName 和 marshal 表示编码的名称和方法；
// requestIDKey 表示 x-request-id 的报头名称，如果为空则表示不需要；
// requestIDGen 表示生成 x-request-id 值的方法；
func NewClient(client *http.Client, codec *Codec, s selector.Selector, marshalName string, marshal func(any) ([]byte, error), requestIDKey string, requestIDGen func() string) *Client {
	if client == nil {
		client = &http.Client{}
	}

	if requestIDKey != "" && requestIDGen == nil {
		panic("当前 requestIDKey 不为空时 requestIDGen 也不能为空")
	}

	return &Client{
		client:   client,
		codec:    codec,
		selector: s,

		marshalName: marshalName,
		marshal:     marshal,

		requestIDKey: requestIDKey,
		requestIDGen: requestIDGen,
	}
}

func (c *Client) Get(path string, resp any, problem *Problem) error {
	return c.Do(http.MethodGet, path, nil, resp, problem)
}

func (c *Client) Delete(path string, resp any, problem *Problem) error {
	return c.Do(http.MethodDelete, path, nil, resp, problem)
}

func (c *Client) Post(path string, req, resp any, problem *Problem) error {
	return c.Do(http.MethodPost, path, req, resp, problem)
}

func (c *Client) Put(path string, req, resp any, problem *Problem) error {
	return c.Do(http.MethodPut, path, req, resp, problem)
}

func (c *Client) Patch(path string, req, resp any, problem *Problem) error {
	return c.Do(http.MethodPatch, path, req, resp, problem)
}

// Do 开始新的请求
//
// req 为提交的对象，最终是由初始化参数的 marshal 进行编码；
// resp 为返回的数据的写入对象，必须是指针类型；
// problem 为返回出错时的写入对象，如果包含自定义的 Extensions 字段，需要为其初始化为零值。
// 非 HTTP 状态码错误返回 err，其它错误由 problem 参数反馈；
func (c *Client) Do(method, path string, req, resp any, problem *Problem) error {
	r, err := c.NewRequest(method, path, req)
	if err != nil {
		return err
	}
	rsp, err := c.Client().Do(r)
	if err != nil {
		return err
	}

	return c.ParseResponse(rsp, resp, problem)
}

// ParseResponse 从 [http.Response] 解析并获取返回对象
//
// 如果正常状态，将内容解码至 resp，如果出错了，则解码至 problem。其它情况下返回错误信息。
func (c *Client) ParseResponse(rsp *http.Response, resp any, problem *Problem) (err error) {
	if rsp.ContentLength == 0 { // 204 可能为空
		return nil
	}

	if problem == nil {
		problem = &Problem{}
	}

	var reader io.Reader = rsp.Body
	encName := rsp.Header.Get(header.ContentEncoding)
	reader, err = c.codec.contentEncoding(encName, reader)
	if err != nil {
		return err
	}

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	if h := rsp.Header.Get(header.ContentType); h != "" {
		if inputMimetype, inputCharset, err = c.codec.contentType(h); err != nil {
			return err
		}

		if inputMimetype == nil {
			return NewLocaleError("not found unmarshaler for the server content-type %s", h)
		}
	} else {
		return NewLocaleError("the server miss content-type header")
	}

	if !header.CharsetIsNop(inputCharset) {
		reader = transform.NewReader(reader, inputCharset.NewDecoder())
	}

	if status.IsProblemStatus(rsp.StatusCode) {
		return inputMimetype(reader, problem)
	}
	return inputMimetype(reader, resp)
}

// NewRequest 生成 [http.Request]
//
// body 为需要提交的对象；
func (c *Client) NewRequest(method, path string, body any) (resp *http.Request, err error) {
	var data []byte
	if body != nil {
		data, err = c.marshal(body)
		if err != nil {
			return nil, err
		}
	}

	if path, err = c.URL(path); err != nil {
		return nil, err
	}

	var r *http.Request
	if len(data) == 0 {
		r, err = http.NewRequest(method, path, nil)
	} else {
		r, err = http.NewRequest(method, path, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, err
	}
	r.Header.Set(header.ContentType, header.BuildContentType(c.marshalName, header.UTF8Name))
	r.Header.Set(header.Accept, c.codec.acceptHeader)
	r.Header.Set(header.AcceptEncoding, c.codec.acceptEncodingHeader)
	if c.requestIDKey != "" {
		r.Header.Set(c.requestIDKey, c.requestIDGen())
	}

	return r, nil
}

// URL 生成一条访问地址
func (c *Client) URL(path string) (string, error) {
	u, err := c.selector.Next()
	if err != nil {
		return "", err
	}
	return u + path, nil
}

func (c *Client) Client() *http.Client { return c.client }
