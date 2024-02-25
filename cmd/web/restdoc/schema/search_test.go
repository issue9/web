// SPDX-License-Identifier: MIT

package schema

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
)

func buildPackages(a *assert.Assertion) *Schema {
	ctx := context.Background()
	l := loggertest.New(a)
	p := New(l.Logger)
	p.Packages().ScanDir(ctx, "./testdata", true)
	return p
}

func TestSchema_New_not_found(t *testing.T) {
	a := assert.New(t, false)
	f := buildPackages(a)
	pkgPath := "github.com/issue9/web/cmd/web/restdoc/schema/testdata"

	// #components/schemas/abc
	t.Run(openapi.ComponentSchemaPrefix, func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		refPath := openapi.ComponentSchemaPrefix + ".admin.notFound"
		ref, err := f.New(context.Background(), tt, refPath, false)
		a.Equal(err, web.NewLocaleError("not found openapi schema ref %s", refPath)).Nil(ref)
	})

	// NotFound
	t.Run("NotFound", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		refPath := pkgPath + "/admin.notFound"
		ref, err := f.New(context.Background(), tt, refPath, false)
		a.Error(err, web.NewLocaleError("not found %s", refPath)).Nil(ref)
	})
}

func TestSchema_New_enum(t *testing.T) {
	a := assert.New(t, false)
	f := buildPackages(a)
	pkgPath := "github.com/issue9/web/cmd/web/restdoc/schema/testdata"
	// pkgRef := refReplacer.Replace(pkgPath)

	t.Run("Enum", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Enum", false)
		a.NotError(err).NotNil(ref).
			NotEmpty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeString).
			Equal(ref.Value.Enum, []any{"v1", "v2", "v3"})
	})

	t.Run("NotBasicTypeEnum", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".NotBasicTypeEnum", false)
		a.Equal(err, web.NewLocaleError("@enum can not be empty")).Nil(ref)
	})

	t.Run("Number", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Number", false)
		a.NotError(err).NotNil(ref).
			NotEmpty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeNumber).
			Equal(ref.Value.Enum, []any{1, 2})
	})
}

func TestSchema_New_types(t *testing.T) {
	a := assert.New(t, false)
	f := buildPackages(a)
	pkgPath := "github.com/issue9/web/cmd/web/restdoc/schema/testdata"
	pkgRef := refReplacer.Replace(pkgPath)

	t.Run("[]bool", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "[]bool", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Value.Type, openapi3.TypeBoolean)
	})

	t.Run("{}", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "{}", false)
		a.NotError(err).Nil(ref)
	})

	t.Run("map", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "map", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeObject)
	})

	t.Run("String", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".String", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Ref). // 基本类型，不需要 Ref
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeString)
	})

	t.Run("Sex", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Sex", false)
		a.NotError(err).NotNil(ref).
			NotEmpty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeString).
			Equal(ref.Value.Enum, []any{"female", "male", "unknown"})
	})

	t.Run("[]String", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "[]"+pkgPath+".String", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Value.Type, openapi3.TypeString)
	})

	// time.Time
	t.Run("time.Time", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "time.Time", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Format, "date-time")
	})

	// time.Duration
	t.Run("time.Duration", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "time.Duration", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeString)
	})

	// 枚举数组
	t.Run("[]Sex", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "[]"+pkgPath+".Sex", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Sex")

		sex, found := tt.GetSchema(pkgRef + ".Sex")
		a.True(found).NotNil(sex).
			Equal(sex.Value.Description, "Sex 表示性别\n\n@enum female male unknown\n@type string\n").
			Equal(sex.Value.Type, "string").
			Equal(sex.Value.Enum, []string{"female", "male", "unknown"})
	})

	t.Run("Sexes", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Sexes", false)
		a.NotError(err).NotNil(ref).
			Empty(ref.Value.Description).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Sex") // Sex 已经被 @type 重定义
	})

	// 对象数组
	t.Run("[]User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, "[]"+pkgPath+".User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeArray).
			Equal(ref.Value.Items.Ref, openapi.ComponentSchemaPrefix+pkgRef+".User")
		u, found := tt.GetSchema(pkgRef + ".User")
		a.True(found).NotNil(u).
			Empty(u.Value.Description). // 单行 doc，赋值给了 title
			Equal(u.Value.Title, "用户信息 doc").
			Equal(u.Value.Type, openapi3.TypeObject)

		name := u.Value.Properties["Name"]
		a.Equal(name.Value.Type, openapi3.TypeString).
			Equal(name.Value.Title, "姓名")

		sex := u.Value.Properties["sex"]
		a.Equal(sex.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".Sex").
			True(sex.Value.XML.Attribute).
			Equal(sex.Value.Title, "性别")

		age := u.Value.Properties["age"]
		a.Empty(age.Ref).
			Equal(age.Value.Title, "年龄").
			Equal(age.Value.Type, openapi3.TypeInteger)

		st := u.Value.Properties["struct"]
		a.Empty(st.Ref).
			Equal(st.Value.Title, "struct doc").
			Equal(st.Value.Type, openapi3.TypeObject)
	})

	// XMLName
	t.Run("XMLName", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.Admin", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.XML.Name, "admin")
	})

	// admin.User
	t.Run("admin.User", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.User", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, openapi.ComponentSchemaPrefix+pkgRef+".admin.User")
		u, found := tt.GetSchema(pkgRef + ".admin.User")
		a.True(found).NotNil(u).
			Equal(u.Value.Title, "User testdata.User").
			Equal(u.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".User")

		name := u.Value.AllOf[0].Value.Properties["Name"]
		a.Equal(name.Value.Type, openapi3.TypeString).
			Equal(name.Value.Title, "姓名", "%+v", name.Value)

		age := u.Value.AllOf[0].Value.Properties["age"]
		a.Empty(age.Ref).
			Equal(age.Value.Title, "年龄").
			Equal(age.Value.Type, openapi3.TypeInteger)
	})

	// admin.Sex
	t.Run("admin.Sex", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.Sex", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Sex")
	})

	// admin.State
	t.Run("admin.State", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.State", false)
		a.Equal(err, web.NewLocaleError("not found type %s", "github.com/issue9/web.State")).Nil(ref)
	})

	// admin.Alias
	t.Run("admin.Alias", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.Alias", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, openapi.ComponentSchemaPrefix+pkgRef+".User")
		u, found := tt.GetSchema(pkgRef + ".User")
		a.True(found).NotNil(u).
			Equal(u.Value.Title, "用户信息 doc")

		name := u.Value.Properties["Name"]
		a.Equal(name.Value.Type, openapi3.TypeString).
			Equal(name.Value.Title, "姓名")

		age := u.Value.Properties["age"]
		a.Empty(age.Ref).
			Equal(age.Value.Title, "年龄").
			Equal(age.Value.Type, openapi3.TypeInteger)
	})

	// admin.Admin
	t.Run("admin.Admin", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.Admin", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Value.Type, openapi3.TypeObject)

		admin, found := tt.GetSchema(pkgRef + ".admin.Admin")
		a.True(found).NotNil(admin)
		name, found := admin.Value.Properties["Name"]
		a.True(found, "%+v", admin.Value.Properties).
			Equal(name.Value.Type, openapi3.TypeString).
			Equal(name.Value.Title, "姓名")

		u1, found := admin.Value.Properties["U1"]
		a.True(found, "%+v", admin.Value.Properties).
			Empty(u1.Ref).
			Equal(u1.Value.Title, "u1").
			Equal(u1.Value.Type, openapi3.TypeArray).
			True(u1.Value.Items.Value.Nullable).
			Equal(u1.Value.Items.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".User")

		u2 := admin.Value.Properties["u2"]
		a.Empty(u2.Ref).
			Equal(u2.Value.Title, "u2").
			True(u2.Value.Nullable).
			Equal(u2.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".User")

		u3, found := admin.Value.Properties["u3"]
		a.False(found).Nil(u3)

		u4 := admin.Value.Properties["U4"]
		a.Equal(u4.Ref, openapi.ComponentSchemaPrefix+pkgRef+".admin.User")
	})
}

func TestSchema_New_generics(t *testing.T) {
	a := assert.New(t, false)
	f := buildPackages(a)
	pkgPath := "github.com/issue9/web/cmd/web/restdoc/schema/testdata"
	pkgRef := refReplacer.Replace(pkgPath)

	// Generic
	t.Run("Generic", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".NumberGenerics[int]", false)
		a.Equal(err, web.NewLocaleError("not found type %s", pkgPath+".NumberGenerics")).
			Nil(ref)
	})

	// Generic[int]
	t.Run("Generic[int]", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Generic[int]", false)
		a.NotError(err).NotNil(ref)
	})

	t.Run("IntGeneric", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".IntGeneric", false)
		a.NotError(err).NotNil(ref)
		v, found := ref.Value.Properties["Type"]
		a.True(found).NotNil(v)
	})

	t.Run("Generics[int,Admin]", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+".Generics[int, "+pkgPath+"/admin.Admin]", false)
		a.NotError(err).NotNil(ref).
			Equal(ref.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Generics-int--"+pkgRef+".admin.Admin-")

		v, found := ref.Value.Properties["F1"]
		a.True(found).NotNil(v)

		v, found = ref.Value.Properties["F2"]
		a.True(found).NotNil(v).
			True(v.Value.Nullable).
			Equal(v.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".admin.Admin")

		v, found = ref.Value.Properties["P"]
		a.True(found).NotNil(v).
			Equal(v.Value.Type, openapi3.TypeInteger)
	})

	t.Run("admin.IntUserGenerics", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.IntUserGenerics", false)
		a.NotError(err).NotNil(ref)

		v, found := ref.Value.Properties["F1"]
		a.True(found).NotNil(v).
			Equal(v.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Generic-int-")

		v, found = ref.Value.Properties["F2"]
		a.True(found).NotNil(v).
			Empty(v.Ref).
			True(v.Value.Nullable).
			Equal(v.Value.AllOf[0].Ref, openapi.ComponentSchemaPrefix+pkgRef+".admin.User")
	})

	t.Run("admin.Int64UserGenerics", func(t *testing.T) {
		a := assert.New(t, false)
		tt := openapi.New("3")

		ref, err := f.New(context.Background(), tt, pkgPath+"/admin.Int64UserGenerics", false)
		a.NotError(err).NotNil(ref)

		v, found := ref.Value.Properties["G1"]
		a.True(found).NotNil(v).
			Equal(v.Ref, openapi.ComponentSchemaPrefix+pkgRef+".Generics-int64--"+pkgRef+".admin.User-")
	})
}
