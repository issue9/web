// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webconfig

import (
	"encoding/xml"
	"io"
)

type pairs map[string]string

type entry struct {
	XMLName struct{} `xml:"key"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

func (p pairs) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(p) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for k, v := range p {
		e.Encode(entry{Name: k, Value: v})
	}

	return e.EncodeToken(start.End())
}

func (p *pairs) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = pairs{}

	for {
		e := &entry{}
		if err := d.Decode(e); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*p)[e.Name] = e.Value
	}

	return nil
}
