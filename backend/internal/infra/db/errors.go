package db

import "errors"

// 业务语义：游标不存在（第一次启动）
var ErrCursorNotFound = errors.New("cursor not found")
