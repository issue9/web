// SPDX-License-Identifier: MIT

package server

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/issue9/sliceutil"
	"github.com/klauspost/compress/zstd"

	"github.com/issue9/web/internal/header"
)

var algWriterPool = sync.Pool{New: func() any { return &algWriter{} }}

type (
	// Encoder 每种压缩实例需要实现的最小接口
	Encoder interface {
		io.WriteCloser
		Reset(io.Writer)
	}

	NewEncoderFunc func() Encoder

	alg struct {
		name string     // 算法名称
		pool *sync.Pool // 算法的对象池

		// contentType 是具体值的，比如 text/xml
		allowTypes []string

		// contentType 是模糊类型的，比如 text/*，
		// 只有在 allowTypes 找不到时，才在此处查找。
		allowTypesPrefix []string
	}

	// 当调用 algWriter.Close 时自动回收到 Pool 中
	algWriter struct {
		Encoder
		b *alg
	}

	compressWriter struct {
		*lzw.Writer
		order lzw.Order
		width int
	}
)

// searchAlg 从报头中查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
func (srv *Server) searchAlg(contentType, h string) (w *alg, notAcceptable bool) {
	if len(srv.algs) == 0 {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
	if len(accepts) == 0 {
		return
	}

	pools := srv.getMatchAlgs(contentType)
	if len(pools) == 0 {
		return
	}

	if last := accepts[len(accepts)-1]; last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, p := range pools {
			exists := sliceutil.Exists(accepts, func(e *header.Item, _ int) bool {
				return e.Value == p.name
			})
			if !exists {
				return p, false
			}
		}
		return
	}

	var identity *header.Item
	for _, accept := range accepts {
		if accept.Err != nil {
			srv.Logs().ERROR().Error(accept.Err)
			continue
		}

		if accept.Value == "identity" { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, a := range pools {
			if a.name == accept.Value {
				return a, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		a := pools[0]
		return a, false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (srv *Server) getMatchAlgs(contentType string) []*alg {
	algs := make([]*alg, 0, len(srv.algs))

LOOP:
	for _, alg := range srv.algs {
		for _, s := range alg.allowTypes {
			if s == contentType {
				algs = append(algs, alg)
				continue LOOP
			}
		}

		for _, p := range alg.allowTypesPrefix {
			if strings.HasPrefix(contentType, p) {
				algs = append(algs, alg)
				continue LOOP
			}
		}
	}

	return algs
}

func (e *algWriter) Close() error {
	err := e.Encoder.Close()
	e.b.pool.Put(e.Encoder)
	algWriterPool.Put(e)
	return err
}

func newAlg(name string, f NewEncoderFunc, ct ...string) *alg {
	types := make([]string, 0, len(ct))
	prefix := make([]string, 0, len(ct))
	for _, c := range ct {
		if c == "" {
			continue
		}

		if c[len(c)-1] == '*' {
			prefix = append(prefix, c[:len(c)-1])
		} else {
			types = append(types, c)
		}
	}
	return &alg{
		name: name,
		pool: &sync.Pool{New: func() any { return f() }},

		allowTypes:       types,
		allowTypesPrefix: prefix,
	}
}

func (p *alg) Get(w io.Writer) io.WriteCloser {
	e := p.pool.Get().(Encoder)
	e.Reset(w)

	aw := algWriterPool.Get().(*algWriter)
	aw.b = p
	aw.Encoder = e
	return aw
}

func (p *alg) Name() string { return p.name }

func (cw *compressWriter) Reset(w io.Writer) {
	cw.Writer.Reset(w, cw.order, cw.width)
}

// GZipWriter 返回指定配置的 gzip 算法
func GZipWriter(level int) NewEncoderFunc {
	return func() Encoder {
		w, err := gzip.NewWriterLevel(nil, level)
		if err != nil {
			panic(err)
		}
		return w
	}
}

// DeflateWriter 返回指定配置的 deflate 算法
func DeflateWriter(level int) NewEncoderFunc {
	return func() Encoder {
		w, err := flate.NewWriter(nil, level)
		if err != nil {
			panic(err)
		}
		return w
	}
}

// BrotliWriter 返回指定配置的 br 算法
func BrotliWriter(o brotli.WriterOptions) NewEncoderFunc {
	return func() Encoder {
		return brotli.NewWriterOptions(nil, o)
	}
}

// CompressWriter 返回指定配置的 compress 算法
func CompressWriter(order lzw.Order, width int) NewEncoderFunc {
	return func() Encoder {
		return &compressWriter{
			Writer: lzw.NewWriter(nil, order, width).(*lzw.Writer),
		}
	}
}

// ZstdWriter 返回指定配置的 zstd 算法
func ZstdWriter(o ...zstd.EOption) NewEncoderFunc {
	return func() Encoder {
		w, err := zstd.NewWriter(nil, o...)
		if err != nil {
			panic(err)
		}
		return w
	}
}
