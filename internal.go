package main

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/15 15:30
* @Package:
 */

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

func parse(fset *token.FileSet, filename string, src []byte, fragmentOk bool) (
	file *ast.File,
	sourceAdj func(src []byte, indent int) []byte,
	indentAdj int,
	err error,
) {
	// Try as whole source file.
	file, err = parser.ParseFile(fset, filename, src, parserMode)

	if err == nil || !fragmentOk || !strings.Contains(err.Error(), "expected 'package'") {
		return
	}

	psrc := append([]byte("package p;"), src...)
	file, err = parser.ParseFile(fset, filename, psrc, parserMode)
	if err == nil {
		sourceAdj = func(src []byte, indent int) []byte {
			// Remove the package clause.
			// Gofmt has turned the ';' into a '\n'.
			src = src[indent+len("package p\n"):]
			return bytes.TrimSpace(src)
		}
		return
	}

	if !strings.Contains(err.Error(), "expected declaration") {
		return
	}

	fsrc := append(append([]byte("package p; func _() {"), src...), '\n', '\n', '}')
	file, err = parser.ParseFile(fset, filename, fsrc, parserMode)
	if err == nil {
		sourceAdj = func(src []byte, indent int) []byte {
			// Cap adjusted indent to zero.
			if indent < 0 {
				indent = 0
			}

			src = src[2*indent+len("package p\n\nfunc _() {"):]
			// Remove only the "}\n" suffix: remaining whitespaces will be trimmed anyway
			src = src[:len(src)-len("}\n")]
			return bytes.TrimSpace(src)
		}

		indentAdj = -1
	}

	// Succeeded, or out of options.
	return
}

func format(
	fset *token.FileSet,
	file *ast.File,
	sourceAdj func(src []byte, indent int) []byte,
	indentAdj int,
	src []byte,
	cfg printer.Config,
) ([]byte, error) {
	if sourceAdj == nil {
		// Complete source file.
		var buf bytes.Buffer
		err := cfg.Fprint(&buf, fset, file)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	i, j := 0, 0
	for j < len(src) && isSpace(src[j]) {
		if src[j] == '\n' {
			i = j + 1 // byte offset of last line in leading space
		}
		j++
	}
	var res []byte
	res = append(res, src[:i]...)

	indent := 0
	hasSpace := false
	for _, b := range src[i:j] {
		switch b {
		case ' ':
			hasSpace = true
		case '\t':
			indent++
		}
	}
	if indent == 0 && hasSpace {
		indent = 1
	}
	for i := 0; i < indent; i++ {
		res = append(res, '\t')
	}

	cfg.Indent = indent + indentAdj
	var buf bytes.Buffer
	err := cfg.Fprint(&buf, fset, file)
	if err != nil {
		return nil, err
	}
	out := sourceAdj(buf.Bytes(), cfg.Indent)

	if len(out) == 0 {
		return src, nil
	}

	// Otherwise, append output to leading space.
	res = append(res, out...)

	// Determine and append trailing space.
	i = len(src)
	for i > 0 && isSpace(src[i-1]) {
		i--
	}
	return append(res, src[i:]...), nil
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
