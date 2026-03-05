package util

import (
	"strconv"
)

// ParseInt64OrDefault parses a string as int64, returning defaultVal on failure.
func ParseInt64OrDefault(s string, defaultVal int64) int64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

// ParseUint64OrDefault parses a string as uint64, returning defaultVal on failure.
func ParseUint64OrDefault(s string, defaultVal uint64) uint64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

// AtoiOrDefault parses a string as int, returning defaultVal on failure.
func AtoiOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

// Page Page
type Page struct {
	PageSize  uint64      `json:"page_size"`
	PageIndex uint64      `json:"page_index"`
	Total     uint64      `json:"total"`
	Data      interface{} `json:"data"`
}

// NewPage NewPage
func NewPage(pageIndex uint64, pageSize uint64, total uint64, data interface{}) *Page {

	return &Page{PageIndex: pageIndex, PageSize: pageSize, Data: data, Total: total}
}

//ToPageNumOrDefault 将字符串转换为数字类型 如果字符串为空 则赋值分页默认参数
func ToPageNumOrDefault(pageIndex string, pageSize string) (pIndex64 uint64, pSize64 uint64) {
	return ParseUint64OrDefault(pageIndex, 1), ParseUint64OrDefault(pageSize, 10)
}
