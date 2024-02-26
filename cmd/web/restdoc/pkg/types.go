// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"github.com/issue9/web"
)

type (
	// Struct 这是对 [types.Struct] 的包装
	Struct struct {
		name   string
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
		tl   typeList
	}

	// [types.TypeList] 并未提供构造方法，用于代替该对象的实例使用。
	typeList interface {
		At(int) types.Type
		Len() int
	}

	// NotFound 表示该类型不存在时返回此类型
	//
	// 一般情况下是引用了未导入的类型，比如：type T = web.State
	// 如果 web 包未被导入，那么 web.State 将会变成 NotFound 类型。
	NotFound string

	// NotImplement 一些未实现的类型
	//
	// 比如 interface{} 作为字段类型时，将返回该对象。
	NotImplement string

	defaultTypeList []types.Type
)

func (s *Struct) String() string { return s.name }

func (s *Struct) Underlying() types.Type { return s }

func (s *Struct) Tag(i int) string { return s.tags[i] }

func (s *Struct) NumFields() int { return len(s.fields) }

func (s *Struct) Field(i int) *types.Var { return s.fields[i] }

// FieldDoc 第 i 个元素的文档
func (s *Struct) FieldDoc(i int) *ast.CommentGroup { return s.docs[i] }

// Doc 关联的文档内容
func (n *Named) Doc() *ast.CommentGroup { return n.doc }

// Next 指向的类型
func (n *Named) Next() types.Type { return n.next }

func (n *Named) TypeArgs() typeList { return n.tl }

// ID 当前对象的唯一名称
func (n *Named) ID() string { return n.id }

func (t NotFound) Underlying() types.Type { return t }

func (t NotFound) String() string { return string(t) }

func (t NotImplement) Underlying() types.Type { return t }

func (t NotImplement) String() string { return string(t) }

func (tl defaultTypeList) At(i int) types.Type { return tl[i] }

func (tl defaultTypeList) Len() int { return len(tl) }

// newTypeList 声明 [typeList] 接口对象
func newTypeList(t ...types.Type) typeList { return defaultTypeList(t) }

// name 为结构体名称，可以为空；
func (pkgs *Packages) newStruct(ctx context.Context, pkg *types.Package, name string, st *ast.StructType, file *ast.File, tl typeList, tps *types.TypeParamList) (*Struct, error) {
	size := st.Fields.NumFields()
	s := &Struct{
		name:   name,
		fields: make([]*types.Var, 0, size),
		docs:   make([]*ast.CommentGroup, 0, size),
		tags:   make([]string, 0, size),
	}

	for _, f := range st.Fields.List {
		doc := getDoc(f.Doc, f.Comment)

		var tag string
		if f.Tag != nil {
			tag = f.Tag.Value
		}

		typ, err := pkgs.typeOfExpr(ctx, pkg, file, f.Type, s, nil, tl, tps)
		if err != nil {
			return nil, err
		}

		switch len(f.Names) {
		case 0:
			s.fields = append(s.fields, types.NewField(f.Pos(), pkg, "", typ, true))
			s.docs = append(s.docs, doc)
			s.tags = append(s.tags, tag)
		case 1:
			s.fields = append(s.fields, types.NewField(f.Pos(), pkg, f.Names[0].Name, typ, false))
			s.docs = append(s.docs, doc)
			s.tags = append(s.tags, tag)
		default:
			for _, n := range f.Names {
				s.fields = append(s.fields, types.NewField(f.Pos(), pkg, n.Name, typ, false))
				s.docs = append(s.docs, doc)
				s.tags = append(s.tags, tag)
			}
		}
	}

	return s, nil
}

// tl 表示范型参数列表，可以为空
func newNamed(named *types.Named, next types.Type, doc *ast.CommentGroup, tl typeList) *Named {
	o := named.Obj()
	id := o.Pkg().Path() + "." + o.Name()

	if named.TypeParams().Len() > 0 && tl != nil && tl.Len() > 0 {
		names := make([]string, 0, tl.Len())
		for i := range tl.Len() {
			names = append(names, tl.At(i).String())
		}
		id += "[" + strings.Join(names, ",") + "]"
	}

	return &Named{
		Named: named,
		next:  next,
		doc:   doc,
		id:    id,
		tl:    tl,
	}
}

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

	wrap, path := splitTypes(path)

	if strings.IndexByte(path, '.') < 0 { // 内置类型
		if t, found := getBasicType(path); found {
			return wrap(t), nil
		}
		return NotFound(path), nil
	}

	// NOTE: 包内可能重定义了内置类型，比如 type int struct {...}
	// 在找不到该类型的情况下，还需尝试将其作为内置类型进行查找。
	var basicType string
	if last := strings.LastIndexByte(path, '.'); last > 0 {
		basicType = path[last+1:]
	}

	typ, err := pkgs.typeOfPath(ctx, path, basicType, nil, tl, nil)
	if err != nil {
		return nil, err
	}
	return wrap(typ), nil
}

// doc 可以为空，参考 typeOfPath；
// self 防止对象引用自身引起的死循环，该值用于判断 expr 是否为与 parent 相同。可以为空；
func (pkgs *Packages) typeOfExpr(ctx context.Context, pkg *types.Package, f *ast.File, expr ast.Expr, self *Struct, doc *ast.CommentGroup, tl typeList, tps *types.TypeParamList) (types.Type, error) {
	switch e := expr.(type) {
	case *ast.SelectorExpr: // type x path.struct
		return pkgs.typeOfPath(ctx, pkgs.getPathFromSelectorExpr(e, f), "", doc, tl, tps)
	case *ast.Ident: // type x y，或是 struct{ f1 T } 中的 T
		if self != nil && e.Name == self.String() {
			return self, nil
		}

		basic := e.Name
		name := pkg.Path() + "." + basic

		if tps != nil && tps.Len() > 0 { // 可能是类型参数名称
			if tl == nil || tl.Len() == 0 {
				return nil, web.NewLocaleError("not found type param %s", e.Name)
			}

			for i := range tps.Len() {
				if tps.At(i).Obj().Name() == e.Name {
					basic = tl.At(i).String()

					if strings.IndexByte(basic, '.') > 0 { // 实参指向 ast.SelectorExpr
						name = basic
						basic = ""
					} else {
						name = pkg.Path() + "." + basic
					}
					break
				}
			}
		}

		return pkgs.typeOfPath(ctx, name, basic, doc, tl, tps)
	case *ast.StructType:
		return pkgs.newStruct(ctx, pkg, "", e, f, tl, tps)
	case *ast.ArrayType: // type x []y
		typ, err := pkgs.typeOfExpr(ctx, pkg, f, e.Elt, self, doc, tl, tps)
		if err != nil {
			return nil, err
		}

		if e.Len == nil {
			return types.NewSlice(typ), nil
		} else {
			l, err := strconv.ParseInt(e.Len.(*ast.BasicLit).Value, 10, 64)
			if err != nil {
				return nil, err
			}
			return types.NewArray(typ, l), nil
		}
	case *ast.StarExpr: // type x *y
		typ, err := pkgs.typeOfExpr(ctx, pkg, f, e.X, self, doc, tl, tps)
		if err != nil {
			return nil, err
		}
		return types.NewPointer(typ), nil
	case *ast.IndexExpr: // type x y[int] 等实例化的范型
		idxType, err := pkgs.typeOfExpr(ctx, pkg, f, e.Index, nil, nil, tl, tps)
		if err != nil {
			return nil, err
		}

		return pkgs.typeOfExpr(ctx, pkg, f, e.X, self, doc, newTypeList(idxType), tps)
	case *ast.IndexListExpr:
		idxTypes := make([]types.Type, 0, len(e.Indices))
		for _, idx := range e.Indices {
			idxType, err := pkgs.typeOfExpr(ctx, pkg, f, idx, nil, nil, tl, tps)
			if err != nil {
				return nil, err
			}
			idxTypes = append(idxTypes, idxType)
		}

		return pkgs.typeOfExpr(ctx, pkg, f, e.X, self, doc, newTypeList(idxTypes...), tps)
	default:
		return NotImplement(""), nil
	}
}

// 获取 path 指向对象的类型
//
// 如果其类型为 [types.Struct]，会被包装为 [Struct]。
// 如果存在类型为 [types.Named]，会被包装为 [Named]。
// 可能存在 type uint string 之类的定义，basicType 表示 path 找不到时是否需要按 basicType 查找基本的内置类型。
// doc 自定义的文档信息，可以为空，表示根据指定的类型信息确定文档。如果是字段类型可以自己指定此值；
func (pkgs *Packages) typeOfPath(ctx context.Context, path, basicType string, doc *ast.CommentGroup, tl typeList, tps *types.TypeParamList) (typ types.Type, err error) {
	obj, spec, f, found := pkgs.lookup(ctx, path)
	if !found {
		if basicType != "" {
			if t, found := getBasicType(basicType); found {
				return t, nil
			}
		}
		return NotFound(path), nil
	}

	if spec == nil { // 只有内置类型此值才可能为 nil，理论上不可能出此错误！
		panic("spec 为空")
	}

	tn, ok := obj.(*types.TypeName)
	if !ok {
		return obj.Type(), nil
	}

	if st, ok := spec.Type.(*ast.StructType); ok {
		typ, err = pkgs.newStruct(ctx, tn.Pkg(), spec.Name.Name, st, f, tl, obj.Type().(*types.Named).TypeParams())
	} else {
		if doc == nil {
			doc = getDoc(spec.Doc, spec.Comment)
		}
		typ, err = pkgs.typeOfExpr(ctx, tn.Pkg(), f, spec.Type, nil, doc, tl, tps)
	}
	if err != nil {
		return nil, err
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		named = types.NewNamed(obj.(*types.TypeName), obj.Type(), nil)
	}
	return newNamed(named, typ, getDoc(spec.Doc, spec.Comment), tl), nil
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
