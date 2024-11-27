// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/issue9/web"
)

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (req *Request) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && req.Ref != nil {
		return nil
	}

	if req.Body != nil {
		if err := req.Body.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Body")
			return err
		}
	}

	for k, v := range req.Content {
		if err := v.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Content[" + k + "]")
			return err
		}
	}

	return nil
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (resp *Response) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && resp.Ref != nil {
		return nil
	}

	for i, p := range resp.Headers {
		if err := p.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Headers[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	if resp.Body != nil {
		if err := resp.Body.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Body")
			return err
		}
	}

	for k, v := range resp.Content {
		if err := v.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Content[" + k + "]")
			return err
		}
	}

	return nil
}

func (s *Server) valid() *web.FieldError {
	params := getPathParams(s.URL)
	if len(params) != len(s.Variables) {
		return web.NewFieldError("Variables", "参数与路径中的指定不同")
	}

	for _, v := range s.Variables {
		if slices.Index(params, v.Name) < 0 {
			return web.NewFieldError("Variables", fmt.Sprintf("参数 %s 不存在于路径中", v.Name))
		}
	}

	return nil
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (t *Parameter) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && t.Ref != nil {
		return nil
	}

	if t.Name == "" {
		return web.NewFieldError("Name", "不能为空")
	}

	if t.Schema == nil {
		return web.NewFieldError("Schema", "不能为空")
	}

	if err := t.Schema.valid(skipRefNotNil); err != nil {
		err.AddFieldParent("Schema")
		return err
	}

	if !t.Schema.isBasicType() {
		return web.NewFieldError("Schema", "不支持复杂类型")
	}

	return nil
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (s *Schema) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && s.Ref != nil {
		return nil
	}

	if s.Type == "" && len(s.AnyOf) == 0 && len(s.AllOf) == 0 && len(s.OneOf) == 0 {
		return web.NewFieldError("Type", "不能为空")
	}

	for i, item := range s.AllOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AllOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	for i, item := range s.AnyOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AnyOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	for i, item := range s.OneOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("OneOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	if s.Items != nil {
		if err := s.Items.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Items")
			return err
		}
		if s.Type != TypeArray {
			return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeArray))
		}
	}
	if s.Type == TypeArray && s.Items == nil {
		return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeArray))
	}

	for key, item := range s.Properties {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Properties[" + key + "]")
			return err
		}
	}
	if len(s.Properties) > 0 && s.Type != TypeObject {
		return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeObject))
	}

	if s.AdditionalProperties != nil {
		if err := s.AdditionalProperties.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AdditionalProperties")
			return err
		}
	}

	return nil
}

func (e *SecurityScheme) valid() *web.FieldError {
	switch e.Type {
	case SecuritySchemeTypeAPIKey:
		if e.Name == "" {
			return web.NewFieldError("Name", "不能为空")
		}
		if e.In != InQuery && e.In != InHeader && e.In != InCookie {
			return web.NewFieldError("In", "无效的值")
		}
	case SecuritySchemeTypeHTTP:
		if e.Scheme == "" {
			return web.NewFieldError("Scheme", "不能为空")
		}
	case SecuritySchemeTypeOAuth2:
		if e.Flows == nil {
			return web.NewFieldError("Flows", "不能为空")
		}
	case SecuritySchemeTypeOpenIDConnect:
		if e.OpenIDConnectURL == "" {
			return web.NewFieldError("OpenIDConnectURL", "不能为空")
		}
	case SecuritySchemeTypeMutualTLS:
	default:
		return web.NewFieldError("Type", "无效的值")
	}
	return nil
}
