// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package pkg 用于对包的解析管理
package pkg

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/issue9/web"
	"golang.org/x/tools/go/packages"

	"github.com/issue9/web/cmd/web/restdoc/logger"
)

const mode = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
	packages.NeedModule | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

const Cancelled = web.StringPhrase("cancelled")

// Packages 管理加载的包
type Packages struct {
	pkgsM sync.Mutex
	pkgs  map[string]*packages.Package // 键名为对应的目录名
	fset  *token.FileSet
	l     *logger.Logger

	// 结构体可能存在相互引用的情况，保存每个结构体的数据，键名为 [Struct.String]。
	structs  map[string]*Struct
	structsM sync.Mutex

	// Packages.TypeOf 会中途加载包文件，比较耗时，
	// 防止 Packages.TypeOf 在执行到一半时又调用此方法加载相同的类型。
	typeOfM sync.Mutex
}

// 根据 st 生成一个空的 [Struct] 或是在已经存在的情况下返回该实例
func (pkgs *Packages) getStruct(st *ast.StructType, tps *types.TypeParamList, tl typeList) (s *Struct, isNew bool) {
	// 根据 st.Pos 与泛型参数形成一个结构的唯一 ID
	id := strconv.Itoa(int(st.Pos()))
	if ss := getTypeParamsList(tps, tl); ss != "" {
		id = id + "[" + ss + "]"
	}

	pkgs.structsM.Lock()
	defer pkgs.structsM.Unlock()

	if s, f := pkgs.structs[id]; f {
		return s, false
	}

	size := st.Fields.NumFields()
	s = &Struct{
		id:     id,
		fields: make([]*types.Var, 0, size),
		docs:   make([]*ast.CommentGroup, 0, size),
		tags:   make([]string, 0, size),
	}
	pkgs.structs[id] = s

	return s, true
}

func New(l *logger.Logger) *Packages {
	return &Packages{
		pkgs: make(map[string]*packages.Package, 30),
		fset: token.NewFileSet(),
		l:    l,

		structs: make(map[string]*Struct, 10),
	}
}

// ScanDir 添加 root 下的内容
//
// root 添加的目录；
func (pkgs *Packages) ScanDir(ctx context.Context, root string, recursive bool) {
	root = filepath.Clean(root)

	pkgs.l.Info(web.Phrase("scan source dir %s", root))

	dirs, err := getDirs(root, recursive)
	if err != nil {
		pkgs.l.Error(err, "", 0)
		return
	}

	wg := &sync.WaitGroup{}
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			pkgs.l.Warning(Cancelled)
			return
		default:
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()

				if _, err := pkgs.load(ctx, dir); err != nil {
					pkgs.l.Error(err, "", 0)
					return
				}
			}(dir)
		}
	}
	wg.Wait()
}

// 加载 dir 目录下的包，如果已经加载，则直接从缓存中读取。
func (pkgs *Packages) load(ctx context.Context, dir string) (*packages.Package, error) {
	dir = filepath.Clean(dir)

	pkgs.pkgsM.Lock()
	defer pkgs.pkgsM.Unlock()
	if p, found := pkgs.pkgs[dir]; found {
		return p, nil
	}

	ps, err := packages.Load(&packages.Config{
		Mode:    mode,
		Context: ctx,
		Dir:     dir,
		Fset:    pkgs.fset,
	})
	if err != nil {
		return nil, err
	}

	switch len(ps) {
	case 0:
		return nil, nil
	case 1:
		pkgs.pkgs[dir] = ps[0]
		return ps[0], nil
	default:
		panic(fmt.Sprintf("目录 %s 中包的数量大于 1：%d", dir, len(ps)))
	}
}

func (pkgs *Packages) FileSet() *token.FileSet { return pkgs.fset }

// Package 返回指定路径的包对象
func (pkgs *Packages) Package(path string) *packages.Package {
	for _, p := range pkgs.pkgs {
		if p.PkgPath == path {
			return p
		}
	}
	return nil
}

// Range 依次访问已经加载的包
//
// f 如果返回了 false，将退出循环
func (pkgs *Packages) Range(f func(*packages.Package) bool) {
	for _, p := range pkgs.pkgs {
		if !f(p) {
			return
		}
	}
}
