package main

import (
	"go/ast"
	"strconv"
	"time"
	"unsafe"
)

/** @Author: zouyx
* @Email:
* @Date:   2024/4/15 16:37
* @Package: 定义字节对齐
 */

const (
	wordSize         uint8 = strconv.IntSize / 8
	leftParenthesis        = "("
	rightParenthesis       = ")"
	leftBrace              = "{"
	rightBrace             = "}"
	structSign             = " struct "
	typeSign               = "type"
	newLine                = '\n'
	markNoSort             = "exclude"
)

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Integer interface {
	Signed | Unsigned
}

type Float interface {
	~float32 | ~float64
}

type Number interface {
	Integer | Float | ~uintptr | ~complex64 | ~complex128
}

type Element interface {
	any
}

func sizeOf[T Element](data T) uint8 {
	return uint8(unsafe.Sizeof(data))
}

var baseFieldSize = map[string]uint8{
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
	"int":        1 * wordSize,
	"uintptr":    1 * wordSize,
	"uint":       1 * wordSize,
	"string":     2 * wordSize,
}

func calculateFieldSize(expr ast.Expr) uint8 {
	var fieldSize uint8
	switch t := expr.(type) {
	case *ast.Ident:
		fieldSize = baseFieldSize[t.Name]
	case *ast.ArrayType:
		fieldSize = sizeOf([]struct{}{})
	case *ast.StarExpr:
		fieldSize = sizeOf(&struct {
		}{})
	case *ast.InterfaceType:
		var tmp interface{}
		fieldSize = sizeOf(tmp)
	case *ast.MapType:
		fieldSize = sizeOf(map[struct{}]struct{}{})
	case *ast.ChanType:
		var tmp chan struct{}
		fieldSize = sizeOf(tmp)
	case *ast.FuncType:
		fieldSize = sizeOf(func() {})
	case *ast.SelectorExpr:
		if t.Sel.Name == "Time" {
			fieldSize = sizeOf(time.Time{})
		}
	default:
		fieldSize = 0
	}
	return fieldSize
}
