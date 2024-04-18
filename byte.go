package main

import (
	"go/ast"
	"syscall"
)

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/15 16:37
* @Package: 定义字节对齐
 */

const (
	wordSize         uint8 = syscall.WORDSIZE / 8
	leftParenthesis        = "("
	rightParenthesis       = ")"
	leftBrace              = "{"
	rightBrace             = "}"
	structSign             = "struct"
	typeSign               = "type"
	newLine                = '\n'
)

var tyeSizeMapping = map[string]uint8{
	"bool":       1,
	"int8":       1,
	"uint8":      1,
	"uint16":     2,
	"int16":      2,
	"int32":      4,
	"uint32":     4,
	"float32":    4,
	"int64":      8,
	"uint64":     8,
	"float64":    8,
	"complex64":  8,
	"complex128": 16,
	"chan":       1 * wordSize,
	"point":      1 * wordSize,
	"map":        1 * wordSize,
	"int":        1 * wordSize,
	"uintptr":    1 * wordSize,
	"func":       1 * wordSize,
	"uint":       1 * wordSize,
	"string":     2 * wordSize,
	"interface":  2 * wordSize,
	"array":      3 * wordSize,
}

// getTypeSize getTypeSize 返回类型占用字节数
func getTypeSize(expr ast.Expr) uint8 {
	var typeStr string
	switch t := expr.(type) {
	case *ast.Ident:
		typeStr = t.Name
	case *ast.ArrayType:
		typeStr = "array"
	case *ast.StarExpr:
		typeStr = "point"
	case *ast.SelectorExpr, *ast.InterfaceType:
		typeStr = "interface"
	case *ast.StructType:
		typeStr = "struct"
	case *ast.MapType:
		typeStr = "map"
	case *ast.ChanType:
		typeStr = "chan"
	case *ast.FuncType:
		typeStr = "func"
	default:
		typeStr = ""
	}
	if size, ok := tyeSizeMapping[typeStr]; ok {
		return size
	}
	return 4 * wordSize
}
