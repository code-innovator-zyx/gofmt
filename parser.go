package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/17 09:56
* @Package: 对struct 按字节对齐的方式优先排序
 */

// 解析struct 返回类型所占总字节数
func parseStruct(structData []byte) ([]byte, uint8) {
	prefix := []byte("package main\n")
	if !bytes.HasPrefix(structData, []byte(typeSign)) {
		prefix = append(prefix, typeSign...)
	}
	builder := bytes.Buffer{}
	builder.Write(prefix)
	builder.Write(structData)
	// 创建一个新的token.FileSet，用来存储位置信息
	fset := token.NewFileSet()
	var (
		res      []byte
		byteSize uint8 = 0
	)

	// 解析源码字符串，返回一个AST
	file, err := parser.ParseFile(fset, "", builder.Bytes(), 0)
	if err != nil {
		panic(err)
	}
	h := sortHeap{
		// 初始化堆
	}
	s := bufio.NewScanner(&builder)
	for s.Scan() {
		res = append(res, append(s.Bytes(), newLine)...)
		if bytes.Contains(s.Bytes(), []byte(structSign)) {
			break
		}
	}
	// 遍历AST节点
	ast.Inspect(file, func(n ast.Node) bool {
		// 查找类型声明节点
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true // 不是类型声明节点，继续遍历
		}
		// 确保类型声明是一个结构体
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true // 不是结构体声明，继续遍历
		}
		// 遍历结构体的字段
		for _, field := range structType.Fields.List {
			for s.Scan() {
				if len(s.Bytes()) != 0 {
					break
				}
			}
			tmp := []byte{}
			tmp = append(tmp, append(s.Bytes(), newLine)...)
			var typeSize uint8 = 0
			if _, ok := field.Type.(*ast.StructType); ok {
				// 返回内置函数 以及函数类型所占总字节数
				tmp, typeSize = innerStruct(s, tmp)
			}
			if typeSize == 0 {
				typeSize = getTypeSize(field.Type)
			}
			//  排除字节对齐的情况下,struct 占用总字节数等于type 占用总字节数相加
			byteSize += typeSize
			//fmt.Printf("++++++\nscore [%d] push  %s \n----------\n", typeSize, string(tmp))
			heap.Push(&h, data{
				typeSize,
				tmp,
			})
		}
		return false // 停止遍历，因为我们已经找到了我们需要的结构体
	})
	heap.Init(&h)
	for h.Len() > 0 {
		res = append(res, heap.Pop(&h).(data).res...)
	}

	for s.Scan() {
		res = append(res, s.Bytes()...)
	}
	return res[len(prefix):], byteSize
}

// 对struct 内置struct进行处理
func innerStruct(scanner *bufio.Scanner, res []byte) ([]byte, uint8) {
	if strings.HasSuffix(scanner.Text(), rightBrace) {
		return parseStruct(res)
	}
	tokenNum := 1
	newLineNum := 1
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			res = append(res, newLine)
			continue
		}
		newLineNum++
		if strings.Contains(scanner.Text(), structSign) && strings.Contains(scanner.Text(), leftBrace) {
			tokenNum++
		}
		if strings.HasSuffix(scanner.Text(), rightBrace) {
			tokenNum--
		}
		if tokenNum >= 0 {
			res = append(res, append(scanner.Bytes(), newLine)...)
		}
		if tokenNum == 0 {
			break
		}
	}
	var byteSize uint8 = 0
	if newLineNum > 1 {
		res, byteSize = parseStruct(res)
	}
	return append(res, newLine), byteSize
}