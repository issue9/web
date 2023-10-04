// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"encoding/xml"
)

func buildMarshalTest(_ *Context) MarshalFunc {
	return func(v any) ([]byte, error) {
		switch vv := v.(type) {
		case error:
			return nil, vv
		default:
			return nil, ErrUnsupportedSerialization()
		}
	}
}

func unmarshalTest(bs []byte, v any) error {
	return ErrUnsupportedSerialization()
}

func marshalJSON(ctx *Context) MarshalFunc { return json.Marshal }

func marshalXML(ctx *Context) MarshalFunc { return xml.Marshal }
