package indexer

import "gorm.io/gorm"

// gormExpr 用于安全地写 SQL 表达式（比如自增）
func gormExpr(expr string) interface{} {
	return gorm.Expr(expr)
}
