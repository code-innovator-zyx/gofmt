// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"golang.org/x/sync/semaphore"
)

var (
	// main operation modes
	list        = flag.Bool("l", false, "list files whose formatting differs from gofmt's")
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	rewriteRule = flag.String("r", "", "rewrite rule (e.g., 'a[b:len(a)] -> a[b:]')")
	simplifyAST = flag.Bool("s", false, "simplify code")
	doDiff      = flag.Bool("d", false, "display diffs instead of rewriting files")
	allErrors   = flag.Bool("e", false, "report all errors (not just the first 10 on different lines)")
	byteAlign   = flag.Bool("a", false, "eorder and optimize all structs through byte alignment")
	// debugging
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to this file")
)

const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers

	printerNormalizeNumbers = 1 << 30
)

var fdSem = make(chan bool, 200)

var (
	rewrite    func(*token.FileSet, *ast.File) *ast.File
	parserMode parser.Mode
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gofmt [flags] [path ...]\n")
	flag.PrintDefaults()
}

func initParserMode() {
	parserMode = parser.ParseComments
	if *allErrors {
		parserMode |= parser.AllErrors
	}

	if *rewriteRule == "" {
		parserMode |= parser.SkipObjectResolution
	}
}

func isGoFile(f fs.DirEntry) bool {
	// ignore non-Go files
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !f.IsDir()
}

type sequencer struct {
	sem       *semaphore.Weighted   // weighted by input bytes (an approximate proxy for memory overhead)
	prev      <-chan *reporterState // 1-buffered
	maxWeight int64
}

func newSequencer(maxWeight int64, out, err io.Writer) *sequencer {
	sem := semaphore.NewWeighted(maxWeight)
	prev := make(chan *reporterState, 1)
	prev <- &reporterState{out: out, err: err}
	return &sequencer{
		maxWeight: maxWeight,
		sem:       sem,
		prev:      prev,
	}
}

const exclusive = -1

func (s *sequencer) Add(weight int64, f func(*reporter) error) {
	if weight < 0 || weight > s.maxWeight {
		weight = s.maxWeight
	}
	if err := s.sem.Acquire(context.TODO(), weight); err != nil {
		// Change the task from "execute f" to "report err".
		weight = 0
		f = func(*reporter) error { return err }
	}

	r := &reporter{prev: s.prev}
	next := make(chan *reporterState, 1)
	s.prev = next

	go func() {
		if err := f(r); err != nil {
			r.Report(err)
		}
		next <- r.getState() // Release the next task.
		s.sem.Release(weight)
	}()
}

func (s *sequencer) AddReport(err error) {
	s.Add(0, func(*reporter) error { return err })
}

func (s *sequencer) GetExitCode() int {
	c := make(chan int, 1)
	s.Add(0, func(r *reporter) error {
		c <- r.ExitCode()
		return nil
	})
	return <-c
}

type reporter struct {
	prev  <-chan *reporterState
	state *reporterState
}

type reporterState struct {
	exitCode int
	//The following fields do not participate in byte alignment sorting. You can make adjustments by yourself
	out, err io.Writer
}

func (r *reporter) getState() *reporterState {
	if r.state == nil {
		r.state = <-r.prev
	}
	return r.state
}

func (r *reporter) Warnf(format string, args ...any) {
	fmt.Fprintf(r.getState().err, format, args...)
}

func (r *reporter) Write(p []byte) (int, error) {
	return r.getState().out.Write(p)
}

func (r *reporter) Report(err error) {
	if err == nil {
		panic("Report with nil error")
	}
	st := r.getState()
	scanner.PrintError(st.err, err)
	st.exitCode = 2
}

func (r *reporter) ExitCode() int {
	return r.getState().exitCode
}

func processFile(filename string, info fs.FileInfo, in io.Reader, r *reporter) error {
	src, err := readFile(filename, info, in)
	if err != nil {
		return err
	}

	fileSet := token.NewFileSet()

	fragmentOk := info == nil
	file, sourceAdj, indentAdj, err := parse(fileSet, filename, src, fragmentOk)
	if err != nil {
		return err
	}

	if rewrite != nil {
		if sourceAdj == nil {
			file = rewrite(fileSet, file)
		} else {
			r.Warnf("warning: rewrite ignored for incomplete programs\n")
		}
	}

	ast.SortImports(fileSet, file)

	if *simplifyAST {
		simplify(file)
	}

	res, err := format(fileSet, file, sourceAdj, indentAdj, src, printer.Config{Mode: printerMode, Tabwidth: tabWidth})
	if err != nil {
		return err
	}
	if !bytes.Equal(src, res) {
		if *byteAlign {
			res = alignStruct(res)
		}
		// formatting has changed
		if *list {
			fmt.Fprintln(r, filename)
		}
		if *write {
			if info == nil {
				panic("-w should not have been allowed with stdin")
			}

			perm := info.Mode().Perm()
			if err := writeFile(filename, src, res, perm, info.Size()); err != nil {
				return err
			}
		}
		if *doDiff {
			newName := filepath.ToSlash(filename)
			oldName := newName + ".orig"
			r.Write(diff(oldName, src, newName, res))
		}
	}

	if !*list && !*write && !*doDiff {
		_, err = r.Write(res)
	}

	return err
}

// 对文件中的struct属性依据其类型进行字节对齐排序
func alignStruct(res []byte) []byte {
	s := bufio.NewScanner(bytes.NewReader(res))
	var (
		sortData = make([]byte, 0, len(res))
	)
	for s.Scan() {
		if len(removeCommentByte(s.Bytes())) == 0 {
			sortData = append(sortData, append(s.Bytes(), newLine)...)
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(s.Text()), "/*") {
			sortData = append(sortData, multiLinComments(s)...)
			continue
		}

		if strings.Contains(removeCommentString(s.Text()), structSign) {
			sortData = append(sortData, append(readGoStruct(s), newLine)...)
			continue
		}
		sortData = append(sortData, append(s.Bytes(), newLine)...)
	}
	return sortData
}

// 多行注释相关
func multiLinComments(scanner *bufio.Scanner) []byte {
	if strings.HasSuffix(strings.TrimSpace(scanner.Text()), "*/") {
		return append(scanner.Bytes(), newLine)
	}
	res := append(scanner.Bytes(), newLine)
	for scanner.Scan() {
		res = append(res, append(scanner.Bytes(), newLine)...)
		if strings.HasSuffix(strings.TrimSpace(scanner.Text()), "*/") {
			break
		}
	}
	return res
}
func readGoStruct(scanner *bufio.Scanner) []byte {
	record := 1
	if strings.HasSuffix(removeCommentString(scanner.Text()), rightBrace) {
		return scanner.Bytes()
	}
	var res []byte
	res = append(res, append(scanner.Bytes(), newLine)...)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		noCommentLine := removeCommentString(scanner.Text())
		if strings.Contains(noCommentLine, leftBrace) {
			record++
		}
		if strings.HasSuffix(noCommentLine, rightBrace) {
			record--
		}
		res = append(res, append(scanner.Bytes(), newLine)...)
		if record == 0 {
			//end
			break
		}
	}
	res, _ = parseStruct(res)
	return res
}

func readFile(filename string, info fs.FileInfo, in io.Reader) ([]byte, error) {
	if in == nil {
		fdSem <- true
		var err error
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		in = f
		defer func() {
			f.Close()
			<-fdSem
		}()
	}

	size := -1
	if info != nil && info.Mode().IsRegular() && int64(int(info.Size())) == info.Size() {
		size = int(info.Size())
	}
	if size+1 <= 0 {
		// The file is not known to be regular, so we don't have a reliable size for it.
		var err error
		src, err := io.ReadAll(in)
		if err != nil {
			return nil, err
		}
		return src, nil
	}

	src := make([]byte, size+1)
	n, err := io.ReadFull(in, src)
	switch err {
	case nil, io.EOF, io.ErrUnexpectedEOF:

	default:
		return nil, err
	}
	if n < size {
		return nil, fmt.Errorf("error: size of %s changed during reading (from %d to %d bytes)", filename, size, n)
	} else if n > size {
		return nil, fmt.Errorf("error: size of %s changed during reading (from %d to >=%d bytes)", filename, size, len(src))
	}
	return src[:n], nil
}

func main() {

	maxWeight := (2 << 20) * int64(runtime.GOMAXPROCS(0))
	s := newSequencer(maxWeight, os.Stdout, os.Stderr)

	gofmtMain(s)
	os.Exit(s.GetExitCode())
}

func gofmtMain(s *sequencer) {
	flag.Usage = usage
	flag.Parse()

	if *cpuprofile != "" {
		fdSem <- true
		f, err := os.Create(*cpuprofile)
		if err != nil {
			s.AddReport(fmt.Errorf("creating cpu profile: %s", err))
			return
		}
		defer func() {
			f.Close()
			<-fdSem
		}()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	initParserMode()
	initRewrite()

	args := flag.Args()
	if len(args) == 0 {
		if *write {
			s.AddReport(fmt.Errorf("error: cannot use -w with standard input"))
			return
		}
		s.Add(0, func(r *reporter) error {
			return processFile("<standard input>", nil, os.Stdin, r)
		})
		return
	}

	for _, arg := range args {
		switch info, err := os.Stat(arg); {
		case err != nil:
			s.AddReport(err)
		case !info.IsDir():
			// Non-directory arguments are always formatted.
			arg := arg
			s.Add(fileWeight(arg, info), func(r *reporter) error {
				return processFile(arg, info, nil, r)
			})
		default:
			// Directories are walked, ignoring non-Go files.
			err := filepath.WalkDir(arg, func(path string, f fs.DirEntry, err error) error {
				if err != nil || !isGoFile(f) {
					return err
				}
				info, err := f.Info()
				if err != nil {
					s.AddReport(err)
					return nil
				}
				s.Add(fileWeight(path, info), func(r *reporter) error {
					return processFile(path, info, nil, r)
				})
				return nil
			})
			if err != nil {
				s.AddReport(err)
			}
		}
	}
}

func fileWeight(path string, info fs.FileInfo) int64 {
	if info == nil {
		return exclusive
	}
	if info.Mode().Type() == fs.ModeSymlink {
		var err error
		info, err = os.Stat(path)
		if err != nil {
			return exclusive
		}
	}
	if !info.Mode().IsRegular() {
		// For non-regular files, FileInfo.Size is system-dependent and thus not a
		// reliable indicator of weight.
		return exclusive
	}
	return info.Size()
}

// writeFile updates a file with the new formatted data.
func writeFile(filename string, orig, formatted []byte, perm fs.FileMode, size int64) error {
	// Make a temporary backup file before rewriting the original file.
	bakname, err := backupFile(filename, orig, perm)
	if err != nil {
		return err
	}

	fdSem <- true
	defer func() { <-fdSem }()

	fout, err := os.OpenFile(filename, os.O_WRONLY, perm)
	if err != nil {
		// We couldn't even open the file, so it should
		// not have changed.
		os.Remove(bakname)
		return err
	}
	defer fout.Close() // for error paths

	restoreFail := func(err error) {
		fmt.Fprintf(os.Stderr, "gofmt: %s: error restoring file to original: %v; backup in %s\n", filename, err, bakname)
	}

	n, err := fout.Write(formatted)
	if err == nil && int64(n) < size {
		err = fout.Truncate(int64(n))
	}

	if err != nil {
		// Rewriting the file failed.

		if n == 0 {
			// Original file unchanged.
			os.Remove(bakname)
			return err
		}

		// Try to restore the original contents.

		no, erro := fout.WriteAt(orig, 0)
		if erro != nil {
			// That failed too.
			restoreFail(erro)
			return err
		}

		if no < n {
			// Original file is shorter. Truncate.
			if erro = fout.Truncate(int64(no)); erro != nil {
				restoreFail(erro)
				return err
			}
		}

		if erro := fout.Close(); erro != nil {
			restoreFail(erro)
			return err
		}

		// Original contents restored.
		os.Remove(bakname)
		return err
	}

	if err := fout.Close(); err != nil {
		restoreFail(err)
		return err
	}

	// File updated.
	os.Remove(bakname)
	return nil
}

func backupFile(filename string, data []byte, perm fs.FileMode) (string, error) {
	fdSem <- true
	defer func() { <-fdSem }()

	nextRandom := func() string {
		return strconv.Itoa(rand.Int())
	}

	dir, base := filepath.Split(filename)
	var (
		bakname string
		f       *os.File
	)
	for {
		bakname = filepath.Join(dir, base+"."+nextRandom())
		var err error
		f, err = os.OpenFile(bakname, os.O_RDWR|os.O_CREATE|os.O_EXCL, perm)
		if err == nil {
			break
		}
		if err != nil && !os.IsExist(err) {
			return "", err
		}
	}

	// write data to backup file
	_, err := f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}

	return bakname, err
}
