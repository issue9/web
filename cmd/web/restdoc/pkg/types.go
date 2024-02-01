// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/issue9/web"
	"golang.org/x/tools/go/packages"
)

type (
	// Struct 这是对 [types.Struct] 的包装
	Struct struct {
		s  *types.Struct
		st *ast.StructType

		fields []*types.Var
		docs   []*ast.CommentGroup
		tags   []string
	}

	// Named 这是对 [types.Named] 的二次包装
	Named struct {
		*types.Named
		next types.Type
		doc  *ast.CommentGroup
		id   string
	}

	// 与 [types.TypeList] 拥有相同的接口，大部分时候也是代替该对象的实例使用。
	typeList interface {
		At(int) types.Type
		Len() int
	}

	// NotFound 表示该类型不存在时返回此类型
	//
	// 一般情况下是引用了未导入的类型，比如：type T = web.State
	// 如果 web 包未被导入，那么 web.State 将会变成 NotFound 类型。
	NotFound string

	defaultTypeList []types.Type
)

func (s *Struct) String() string { return s.s.String() }

func (s *Struct) Underlying() types.Type { return s.s.Underlying() }

func (s *Struct) Tag(i int) string { return s.tags[i] }

func (s *Struct) NumFields() int { return len(s.fields) }

func (s *Struct) Field(i int) *types.Var { return s.fields[i] }

// FieldDoc 第 i 个元素的文档
func (s *Struct) FieldDoc(i int) *ast.CommentGroup { return s.docs[i] }

// Doc 关联的文档内容
func (n *Named) Doc() *ast.CommentGroup { return n.doc }

// Next 指向的类型
func (n *Named) Next() types.Type { return n.next }

// ID 当前对象的唯一名称
func (n *Named) ID() string { return n.id }

func (t NotFound) Underlying() types.Type { return t }

func (t NotFound) String() string { return string(t) }

func (tl defaultTypeList) At(i int) types.Type { return tl[i] }

func (tl defaultTypeList) Len() int { return len(tl) }

// newTypeList 声明 [TypeList] 接口对象
func newTypeList(t ...types.Type) typeList { return defaultTypeList(t) }

// TypeOf 查找名为 path 的相关类型信息
//
// path 为完整的类型名，需要包含路径部分。完整格式如下：
//
//	[prefix][path.]type[[type param]]
//
// 其中 prefix 表示类型修改的前缀，可以有以下三种格式：
//   - [] 表示数组；
//   - * 表示指针；
//   - [x] 数组，x 必须得是正整数；
//
// path 表示类型的包路径，如果是非内置类型，该值是必须的；
// type param 表示泛型的实参，比如 [int, float] 等；
// path 拥有以下两个特殊值：
//   - {} 表示空值，将返回 nil, true
//   - map 或是 any 将返回 [types.InterfaceType]
func (pkgs *Packages) TypeOf(ctx context.Context, path string) (types.Type, error) {
	path, tl, err := pkgs.splitTypeParams(ctx, path)
	if err != nil {
		return nil, err
	}
	return pkgs.typeOf(ctx, path, tl)
}

func (pkgs *Packages) typeOf(ctx context.Context, path string, tl typeList) (types.Type, error) {
	f, path := splitTypes(path)

	if strings.IndexByte(path, '.') < 0 { // 内置类型
		if t, found := getBasicType(path); found {
			return f(t), nil
		}
		return nil, web.NewLocaleError("%s is not a valid basic type", path)
	}

	obj, spec, file, found := pkgs.lookup(ctx, path)
	if !found {
		// NOTE: 包内可能重定义了内置类型，比如 type int struct {...}
		// 在找不到该类型的情况下，还需尝试将其作为内置类型进行查找。
		if last := strings.LastIndexByte(path, '.'); last > 0 {
			if t, found := getBasicType(path[last+1:]); found {
				return f(t), nil
			}
		}
		return NotFound(path), nil
	}

	typ, err := pkgs.typeOfExpr(ctx, obj, spec.Type, file, getDoc(spec.Doc, spec.Comment), tl)
	if err != nil {
		return nil, err
	}

	return f(typ), nil
}

func (pkgs *Packages) newStruct(ctx context.Context, ts *types.Struct, st *ast.StructType, tl typeList, tps *types.TypeParamList) (*Struct, error) {
	s := &Struct{
		s:      ts,
		st:     st,
		fields: make([]*types.Var, 0, ts.NumFields()),
		docs:   make([]*ast.CommentGroup, 0, ts.NumFields()),
		tags:   make([]string, 0, ts.NumFields()),
	}

	// BUG 结构体中的字段如果引用自身，会造成死循环

	if err := pkgs.addField(ctx, s, ts, st, tl, tps); err != nil {
		return nil, err
	}
	return s, nil
}

// tl 表示范型参数列表可以为空
func newNamed(named *types.Named, next types.Type, doc *ast.CommentGroup, tl typeList) *Named {
	o := named.Obj()
	id := o.Pkg().Path() + "." + o.Name()
	if named.TypeParams().Len() > 0 {
		names := make([]string, 0, tl.Len())
		for i := 0; i < tl.Len(); i++ {
			names = append(names, tl.At(i).String()) // BUG 如果 tl.At(i) 是个嵌套的范型
		}
		id = id + "[" + strings.Join(names, ",") + "]"
	}

	return &Named{
		Named: named,
		next:  next,
		doc:   doc,
		id:    id,
	}
}

// 将 ts 的所有有字段加入 s 之中
func (pkgs *Packages) addField(ctx context.Context, s *Struct, ts *types.Struct, st *ast.StructType, tl typeList, tps *types.TypeParamList) error {
	for i := 0; i < ts.NumFields(); i++ {
		v := ts.Field(i)

		if v.Anonymous() {
			tt, err := pkgs.typeOfFieldType(ctx, v.Type(), tl, tps)
			if err != nil {
				return err
			}

			named, ok := tt.(*Named)
			if ok {
				tt = named.Next()
			}

			switch t := tt.(type) {
			case *Struct:
				if err := pkgs.addField(ctx, s, t.s, t.st, tl, tps); err != nil {
					return err
				}
			default:
				s.fields = append(s.fields, types.NewVar(v.Pos(), v.Pkg(), v.Name(), tt))
				s.docs = append(s.docs, nil)
				s.tags = append(s.tags, "")
			}

			continue
		}

		if !v.Exported() {
			continue
		}

		switch t := v.Type().(type) {
		case *types.Struct: // 结构体中的结构体
			ps, err := pkgs.newStruct(ctx, t, st.Fields.List[i].Type.(*ast.StructType), tl, tps)
			if err != nil {
				return err
			}
			v = types.NewVar(v.Pos(), v.Pkg(), v.Name(), ps)
		default:
			tt, err := pkgs.typeOfFieldType(ctx, t, tl, tps)
			if err != nil {
				return err
			}
			v = types.NewVar(v.Pos(), v.Pkg(), v.Name(), tt)
		}

		s.fields = append(s.fields, v)
		s.docs = append(s.docs, getDoc(st.Fields.List[i].Doc, st.Fields.List[i].Comment))
		var tag string
		if st.Fields.List[i].Tag != nil {
			tag = st.Fields.List[i].Tag.Value
		}
		s.tags = append(s.tags, tag)
	}

	return nil
}

// 获取结构体字段的类型
func (pkgs *Packages) typeOfFieldType(ctx context.Context, typ types.Type, tl typeList, tps *types.TypeParamList) (t types.Type, err error) {
	switch tt := typ.(type) {
	case *types.Pointer:
		t, err = pkgs.typeOfFieldType(ctx, tt.Elem(), tl, tps)
		if err == nil {
			t = types.NewPointer(t)
		}
	case *types.Array:
		t, err = pkgs.typeOfFieldType(ctx, tt.Elem(), tl, tps)
		if err == nil {
			t = types.NewArray(t, tt.Len())
		}
	case *types.Slice:
		t, err = pkgs.typeOfFieldType(ctx, tt.Elem(), tl, tps)
		if err == nil {
			t = types.NewSlice(t)
		}
	case *types.TypeParam:
		if tl != nil && tl.Len() > 0 {
			t, err = pkgs.typeOfFieldType(ctx, tl.At(tt.Index()), tl, tps)
		} else {
			return nil, web.NewLocaleError("the type %s unset type params", tt.Obj().Name())
		}
	case *types.Named:
		ts := make([]types.Type, 0, tt.TypeParams().Len())
		for i := 0; i < tt.TypeArgs().Len(); i++ {
			t := getFieldTypeParam(tps, tl, tt.TypeArgs().At(i).String())
			if t != nil {
				ts = append(ts, t)
			}
		}
		path := tt.String()
		if index := strings.IndexByte(path, '['); index > 0 {
			path = path[:index]
		}
		t, err = pkgs.typeOf(ctx, path, newTypeList(ts...))
	default: // 其它类型原样返回
		t, err = pkgs.typeOf(ctx, tt.String(), tl)
	}
	return
}

func (pkgs *Packages) typeOfExpr(ctx context.Context, obj types.Object, expr ast.Expr, file *ast.File, doc *ast.CommentGroup, tl typeList) (typ types.Type, err error) {
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return obj.Type(), nil
	}
	pkgPath := tn.Pkg().Path()

	var xnamed *types.Named
	if named, ok := obj.Type().(*types.Named); ok {
		xnamed = named
	} else {
		xnamed = types.NewNamed(tn, obj.Type(), nil)
	}

	switch e := expr.(type) {
	case *ast.SelectorExpr: // type x path.struct
		typ, err = pkgs.typeOfSelectorExpr(ctx, e, file, tl)
	case *ast.Ident: // type x y
		typ, err = pkgs.typeOfIdent(ctx, pkgPath, e, tl)
	case *ast.ArrayType: // type x []y
		switch elt := e.Elt.(type) {
		case *ast.SelectorExpr:
			typ, err = pkgs.typeOfSelectorExpr(ctx, elt, file, tl)
		case *ast.Ident:
			typ, err = pkgs.typeOfIdent(ctx, pkgPath, elt, tl)
		default:
			panic(fmt.Sprintf("未处理的 ast.ArrayType.Elt 类型： %T", expr))
		}
		if err != nil {
			return nil, err
		}
		if e.Len == nil {
			typ = types.NewSlice(typ)
		} else {
			l, err := strconv.ParseInt(e.Len.(*ast.BasicLit).Value, 10, 64)
			if err != nil {
				return nil, err
			}
			typ = types.NewArray(typ, l)
		}
	case *ast.StarExpr: // type x *y
		switch x := e.X.(type) {
		case *ast.SelectorExpr:
			typ, err = pkgs.typeOfSelectorExpr(ctx, x, file, tl)
		case *ast.Ident:
			typ, err = pkgs.typeOfIdent(ctx, pkgPath, x, tl)
		default:
			panic(fmt.Sprintf("未处理的 ast.StartExpr.X 类型： %T", expr))
		}

		if err != nil {
			return nil, err
		}
		typ = types.NewPointer(typ)
	case *ast.StructType:
		switch {
		// 如果是范型且指定了类型参数，那么就用其指定的类型参数, 比如 type x = y[int] 等
		case xnamed.TypeArgs().Len() == xnamed.TypeParams().Len():
			tl = xnamed.TypeArgs()
		// 由外界指定类型参数，比如 pkgs.typeOf("github.com/issue9/pkg.G", NewTypeList(int))
		case tl != nil && xnamed.TypeParams().Len() == tl.Len():
		default:
			return nil, web.NewLocaleError("uninstance type %s", obj.Type().String())
		}

		typ, err = pkgs.newStruct(ctx, obj.Type().Underlying().(*types.Struct), e, tl, xnamed.TypeParams())
	case *ast.IndexExpr: // type x y[int] 等实例化的范型
		var idxType types.Type
		switch idx := e.Index.(type) {
		case *ast.Ident:
			idxType, err = pkgs.typeOfIdent(ctx, pkgPath, idx, nil)
		case *ast.SelectorExpr:
			idxType, err = pkgs.typeOfSelectorExpr(ctx, idx, file, nil)
		}

		if err != nil {
			return nil, err
		}

		return pkgs.typeOfExpr(ctx, xnamed.Obj(), e.X, file, doc, newTypeList(idxType))
	case *ast.IndexListExpr:
		tps := make([]types.Type, 0, len(e.Indices))
		for _, idx := range e.Indices {
			var idxType types.Type
			switch expr := idx.(type) {
			case *ast.Ident:
				idxType, err = pkgs.typeOfIdent(ctx, pkgPath, expr, nil)
			case *ast.SelectorExpr:
				idxType, err = pkgs.typeOfSelectorExpr(ctx, expr, file, nil)
			}

			if err != nil {
				return nil, err
			}
			tps = append(tps, idxType)
		}

		return pkgs.typeOfExpr(ctx, xnamed.Obj(), e.X, file, doc, newTypeList(tps...))
	case *ast.InterfaceType:
		return nil, web.NewLocaleError("ast.InterfaceType can not covert to openapi schema", expr)
	default:
		panic(fmt.Sprintf("未处理的 ast.Expr 类型： %T", expr))
	}

	if err != nil {
		return nil, err
	}

	return newNamed(xnamed, typ, doc, tl), nil
}

func (pkgs *Packages) typeOfSelectorExpr(ctx context.Context, expr *ast.SelectorExpr, f *ast.File, tl typeList) (types.Type, error) {
	// BUG expr 可能包含范型
	return pkgs.typeOfPath(ctx, pkgs.getPathFromSelectorExpr(expr, f), "", tl)
}

func (pkgs *Packages) typeOfIdent(ctx context.Context, currPkg string, ident *ast.Ident, tl typeList) (types.Type, error) {
	return pkgs.typeOfPath(ctx, currPkg+"."+ident.Name, ident.Name, tl)
}

// 获取 path 指向对象的类型
//
// 如果其类型为 [types.Struct]，会被包装为 [Struct]。
// 如果存在类型为 [types.Named]，会被包装为 [Named]。
// 可能存在 type uint string 之类的定义，basicType 表示 path 找不到时是否需要按 basicType 查找基本的内置类型。
func (pkgs *Packages) typeOfPath(ctx context.Context, path, basicType string, tl typeList) (typ types.Type, err error) {
	obj, spec, f, found := pkgs.lookup(ctx, path)
	if !found {
		if basicType != "" {
			if t, found := getBasicType(basicType); found {
				return t, nil
			}
			return nil, web.NewLocaleError("%s is not a valid basic type", basicType)
		}
		return NotFound(path), nil
	}

	if spec == nil { // 只有内置类型此值才可能为 nil，理论上不可能出此错误！
		panic("spec 为空")
	}

	switch t := spec.Type.(type) {
	case *ast.SelectorExpr:
		typ, err = pkgs.typeOfPath(ctx, pkgs.getPathFromSelectorExpr(t, f), "", tl)
	case *ast.Ident: // NOTE: t.Name 可能是基础类型
		typ, err = pkgs.typeOfPath(ctx, obj.Pkg().Path()+"."+t.Name, t.Name, tl)
	case *ast.StructType:
		typ, err = pkgs.newStruct(ctx, obj.Type().Underlying().(*types.Struct), t, tl, obj.Type().(*types.Named).TypeParams())
	default:
		panic(fmt.Sprintf("未处理的 ast.TypeSpec.Type 类型： %T", spec.Type))
	}

	if err != nil {
		return nil, err
	}

	var xnamed *types.Named
	if basic, ok := typ.(*types.Basic); ok { // 基础类型，则表示没有更深层的内容。
		xnamed = types.NewNamed(obj.(*types.TypeName), basic, nil)
	} else {
		xnamed = obj.Type().(*types.Named)
	}
	return newNamed(xnamed, typ, getDoc(spec.Doc, spec.Comment), tl), nil
}

func findInPkgs(ps []*packages.Package, pkgPath, typeName string) (types.Object, *ast.TypeSpec, *ast.File, bool) {
	for _, p := range ps {
		if p.PkgPath != pkgPath {
			continue
		}

		obj := p.Types.Scope().Lookup(typeName)
		if obj == nil {
			break // 不可能存在多个 path 相同但内容不同的 Package 对象
		}

		for _, f := range p.Syntax {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != typeName {
						continue
					}

					// 整个 type() 范围只有一个类型，直接采用 type 的注释
					if ts.Doc == nil && len(gen.Specs) == 1 {
						ts.Doc = gen.Doc
					}
					return obj, ts, f, true
				}
			}
		}
	}

	return nil, nil, nil, false
}

func getBasicType(name string) (types.Type, bool) {
	for _, b := range types.Typ {
		if b.Name() == name {
			return b, true
		}
	}

	switch name {
	case "map", "any", "interface{}":
		return types.NewInterfaceType(nil, nil), true
	case "{}":
		return nil, true
	}

	return nil, false
}

func getDoc(doc, comment *ast.CommentGroup) *ast.CommentGroup {
	if doc == nil {
		doc = comment
	}
	return doc
}

func getFieldTypeParam(tps *types.TypeParamList, tl typeList, name string) types.Type {
	for i := 0; i < tps.Len(); i++ {
		tp := tps.At(i)
		if tp.Obj().Name() == name {
			return tl.At(tp.Index())
		}
	}
	return nil
}
