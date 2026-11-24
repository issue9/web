// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation

import (
	"fmt"
	"io"
	"iter"
	"net/http"
	"slices"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/qheader"
)

// Codec 媒体类型需要提供的编码接口
//
// 实现者应该是一个无状态的对象。
type Codec[T any] interface {
	// Marshal 对 v 进行编码
	Marshal(ctx T, v any) ([]byte, error)

	// Unmarshal 解码内容到 v
	Unmarshal(r io.Reader, v any) error
}

// MediaType 媒体类型
type MediaType[T any] struct {
	Codec[T]
	name, problem, method string
	kind                  PatchKind
}

type PatchKind = int8

const (
	PatchKindNone PatchKind = iota
	PatchKindPatcher
	PatchKindMerger
)

// NewMediaType 创建新的媒体类型
//
// name 媒体类型的名称；
// problem 错误状态下的媒体类型；
// method 绑定的方法，如果该值不为空那么 problem 参数将无效；
// pathKind 在 method 的值为 [http.MethodPatch] 时可以使用此值表示可解析的对象类型；
//
// method 有以下几种请求方法有：
//   - 空值，可用于处理所有的类型；
//   - [http.MethodPatch] 只能处理由 PATCH 请求提交的数据；
//   - [http.MethodQuery] 只能处理由 QUERY 请求提交的数据；
func NewMediaType[T any](codec Codec[T], name, problem, method string, patchKind PatchKind) *MediaType[T] {
	if codec == nil {
		panic("参数 codec 不能为 nil")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}
	if problem == "" {
		problem = name
	}

	if method != http.MethodPatch && patchKind != PatchKindNone {
		panic("patchKind 只在 method 为 PATCH 时有效")
	}

	return &MediaType[T]{
		Codec:   codec,
		name:    name,
		problem: problem,
		method:  method,
		kind:    patchKind,
	}
}

func (mt *MediaType[T]) Name() string         { return mt.name }
func (mt *MediaType[T]) ProblemName() string  { return mt.problem }
func (mt *MediaType[T]) Method() string       { return mt.method }
func (mt *MediaType[T]) PatchKind() PatchKind { return mt.kind }

// AddMediaType 添加媒体类型
func (nego *Negotiation[T]) AddMediaType(mt ...*MediaType[T]) *Negotiation[T] {
	for _, mm := range mt {
		if slices.ContainsFunc(nego.mediaTypes, func(v *MediaType[T]) bool { return v.Name() == mm.Name() }) {
			panic(fmt.Sprintf("存在重复的项 %s", mm.Name()))
		}

		nego.mediaTypes = append(nego.mediaTypes, mm)

		switch m := mm.Method(); m {
		case "":
			if nego.acceptHeader != "" {
				nego.acceptHeader += ", "
			}
			nego.acceptHeader += mm.Name()
		case http.MethodPatch:
			if nego.acceptPatchHeader != "" {
				nego.acceptPatchHeader += ", "
			}
			nego.acceptPatchHeader += mm.Name()
		case "QUERY":
			if nego.acceptQueryHeader != "" {
				nego.acceptQueryHeader += ", "
			}
			nego.acceptQueryHeader += mm.Name()
		}
	}

	return nego
}

// AddCodec 添加编码
func (nego *Negotiation[T]) AddCodec(codec Codec[T], name, problem, method string, kind PatchKind) *Negotiation[T] {
	return nego.AddMediaType(NewMediaType(codec, name, problem, method, kind))
}

func unableParse(v string) error { return localeutil.Error("unable parse %s", v) }

func (nego *Negotiation[T]) AcceptHeader() string { return nego.acceptHeader }

func (nego *Negotiation[T]) AcceptQueryHeader() string { return nego.acceptQueryHeader }

func (nego *Negotiation[T]) AcceptPatchHeader() string { return nego.acceptPatchHeader }

// Accept 从 Accept 报头解析出所需要的解码函数
//
// 如果 h 为空，那么将被设置为默认值 ＊/*；
func (nego *Negotiation[T]) Accept(h string) (*MediaType[T], error) {
	// 规则如下：
	//  - 按 q 从大到小进行排序；
	//  - 忽略无法解析 q 的项；
	//  - 依次匹配，匹配到即返回；
	//  - 如果匹配到的项是 */*，那么还要查看匹配到的项是否在 q=0 的项中，如果是则匹配下一个项；
	//
	// 如果传递的内容如下：
	//
	//	application/json;q=0.9,*/*;q=1
	//
	// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
	//
	// */*;q=0 表示匹配所有未列出的编码

	accepts, deny, errs := qheader.ParseQHeader(h, "*/*")
	defer qheader.FreeItems(&accepts)

	if len(accepts) == 0 && len(deny) == 0 && len(errs) > 0 {
		return nil, unableParse(h)
	}

	hasAny := false
	for _, item := range accepts {
		if item.Value == "*/*" { // */* 必然在切片的最后
			hasAny = true
			break
		}

		for _, mt := range nego.mediaTypes {
			name, mn := item.Value, mt.Name()
			if name == mn || matchPrefix(mn, name) {
				return mt, nil
			}
		}
	}

	// 在 accept 中未找到符合的，且指定 */*，挑一个服务器支持且不在 deny 中的
	if hasAny {
		for _, mt := range nego.mediaTypes {
			if !mediaTypeIsDenied(mt.Name(), deny) {
				return mt, nil
			}
		}
	}

	return nil, nil
}

// name 是否在 deny 中
func mediaTypeIsDenied(name string, deny []string) bool {
	for _, d := range deny {
		if d == "*/*" || d == name || matchPrefix(name, d) {
			return true
		}
	}
	return false
}

// match 为 XX/* 且 v 符合 match 的规则
func matchPrefix(v, match string) bool {
	return strings.HasSuffix(match, "/*") && strings.HasPrefix(v, match[:len(match)-1])
}

// ContentType 从 Content-Type 报头中获取解码和字符集函数
//
// h 表示 Content-Type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
// method 如果为空表示只从 [MediaType.Method] 为空的类型中查找，
// 如果不为空，除了前者还会从 [MediaType.Method] 与 method 相同的类型中查找；
func (nego *Negotiation[T]) ContentType(h, method string) (*MediaType[T], encoding.Encoding, error) {
	mt, charset := qheader.ParseWithParam(h, "charset")

	item, found := sliceutil.At(nego.mediaTypes, func(i *MediaType[T], _ int) bool {
		return (i.Name() == mt) && (method == "" || i.Method() == "" || i.Method() == method)
	})
	if !found {
		return nil, nil, localeutil.Error("not found serialization function for %s", mt)
	}

	if charset == "" || charset == header.UTF8 {
		return item, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return item, c, nil
}

// MediaTypes 返回支持的所有媒体类型
func (nego *Negotiation[T]) MediaTypes() iter.Seq[*MediaType[T]] {
	return func(yield func(*MediaType[T]) bool) {
		for _, v := range nego.mediaTypes {
			if !yield(v) {
				break
			}
		}
	}
}
