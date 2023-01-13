// SPDX-License-Identifier: MIT

package cache

// 常用时间的定义
const (
	Forever     = 0 //  永不过时
	OneMinute   = 60
	FiveMinutes = 5 * OneMinute
	TenMinutes  = 10 * OneMinute
	HalfHour    = 30 * OneMinute
	OneHour     = 60 * OneMinute
	HalfDay     = 12 * OneHour
	OneDay      = 24 * OneHour
	OneWeek     = 7 * OneDay
	ThirtyDays  = 30 * OneDay // 30 天
	SixtyDays   = 60 * OneDay // 60 天
	NinetyDays  = 90 * OneDay // 90 天
)
