// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package app

import "github.com/kardianos/service"

type empty struct{}

var _ service.Interface = &app{}
