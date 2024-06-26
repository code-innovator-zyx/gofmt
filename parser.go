package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strings"
)

/** @Author: zouyx
* @Email:
* @Date:   2024/4/17 09:56
* @Package: 对struct 按字节对齐的方式优先排序
 */

// 解析struct 返回类型所占总字节数
func parseStruct(structData []byte) ([]byte, uint16) {
	prefix := []byte("package main\n")
	if !bytes.HasPrefix(removeCommentByte(bytes.TrimSpace(structData)), []byte(typeSign)) {
		prefix = append(prefix, typeSign...)
	}
	builder := bytes.Buffer{}
	builder.Write(prefix)
	builder.Write(structData)
	// 创建一个新的token.FileSet，用来存储位置信息
	fset := token.NewFileSet()
	var (
		res      []byte
		byteSize uint16 = 0
	)
	//fmt.Println(builder.String())
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
		if strings.Contains(removeCommentString(s.Text()), structSign) {
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
				if len(removeCommentString(s.Text())) != 0 {
					break
				}
			}
			tmp := []byte{}
			tmp = append(tmp, append(s.Bytes(), newLine)...)
			var typeSize uint16 = 0
			if _, ok := field.Type.(*ast.StructType); ok {
				// 返回内置函数 以及函数类型所占总字节数
				tmp, typeSize = innerStruct(s, tmp)
			}
			if typeSize == 0 {
				typeSize = uint16(calculateFieldSize(field.Type))
			}

			//  排除字节对齐的情况下,struct 占用总字节数等于type 占用总字节数相加
			byteSize += typeSize

			heap.Push(&h, data{
				score: typeSize,
				res:   tmp,
			})
		}
		return false // 停止遍历，因为我们已经找到了我们需要的结构体
	})
	heap.Init(&h)
	var hasMark bool
	for h.Len() > 0 {
		d := heap.Pop(&h).(data)
		if d.score == 0 && hasMark == false {
			res = append(res, []byte("	//The following fields do not participate in byte alignment sorting. You can make adjustments by yourself\n")...)
			hasMark = true
		}
		res = append(res, d.res...)
	}

	for s.Scan() {
		res = append(res, append(s.Bytes(), newLine)...)
	}
	// 如果是一个空的struct，那么这个struct理应放在首行
	if byteSize == 0 {
		byteSize = math.MaxUint16
	}
	return res[len(prefix):], byteSize
}

// 对struct 内置struct进行处理
func innerStruct(scanner *bufio.Scanner, res []byte) ([]byte, uint16) {
	if strings.HasSuffix(removeCommentString(scanner.Text()), rightBrace) {
		return parseStruct(res)
	}
	tokenNum := 1
	newLineNum := 1
	var suffix []byte
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			res = append(res, newLine)
			continue
		}
		newLineNum++
		noCommentLine := removeCommentString(scanner.Text())
		if strings.Contains(noCommentLine, leftBrace) {
			tokenNum++
		}
		if strings.HasSuffix(noCommentLine, rightBrace) {
			tokenNum--
		}
		if tokenNum > 0 {
			res = append(res, append(scanner.Bytes(), newLine)...)
		}
		if tokenNum == 0 {
			lastLine := scanner.Bytes()
			// inner struct may has `json: ` format
			if index := bytes.Index(removeCommentByte(lastLine), []byte("`")); index != -1 {
				suffix = lastLine[index:]
				lastLine = lastLine[:index]
			}
			res = append(res, lastLine...)
			break
		}
	}
	var byteSize uint16 = 0
	if newLineNum > 1 {
		res, byteSize = parseStruct(res)
		res = res[:bytes.LastIndexByte(res, newLine)]
		res = append(res, suffix...)
	}
	return append(res, newLine), byteSize
}
