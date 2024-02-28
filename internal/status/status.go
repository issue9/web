// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate go run ./make_statuses.go

// Package status 用于处理与状态码相关的功能
package status

func IsProblemStatus(status int) bool {
	_, found := problemStatuses[status]
	return found
}
