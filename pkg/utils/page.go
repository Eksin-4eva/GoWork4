package utils

import "gobili/pkg/constants"

// NormalizePage 兜底修正分页参数。
//
// 客户端传 0、负数或超大 page_size 都不该让数据库去承担，
// 因此统一在这里夹到合法区间。
func NormalizePage(pageNum, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = constants.DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	return pageNum, pageSize
}

// Offset 由页码换算 SQL OFFSET。
func Offset(pageNum, pageSize int) int {
	return (pageNum - 1) * pageSize
}
