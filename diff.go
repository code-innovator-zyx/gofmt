package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/15 15:36
* @Package:
 */

type pair struct{ x, y int }

func diff(oldName string, old []byte, newName string, new []byte) []byte {
	if bytes.Equal(old, new) {
		return nil
	}
	x := lines(old)
	y := lines(new)

	// Print diff header.
	var out bytes.Buffer
	fmt.Fprintf(&out, "diff %s %s\n", oldName, newName)
	fmt.Fprintf(&out, "--- %s\n", oldName)
	fmt.Fprintf(&out, "+++ %s\n", newName)

	var (
		done  pair     // printed up to x[:done.x] and y[:done.y]
		chunk pair     // start lines of current chunk
		count pair     // number of lines from each side in current chunk
		ctext []string // lines for current chunk
	)
	for _, m := range tgs(x, y) {
		if m.x < done.x {
			// Already handled scanning forward from earlier match.
			continue
		}

		start := m
		for start.x > done.x && start.y > done.y && x[start.x-1] == y[start.y-1] {
			start.x--
			start.y--
		}
		end := m
		for end.x < len(x) && end.y < len(y) && x[end.x] == y[end.y] {
			end.x++
			end.y++
		}

		for _, s := range x[done.x:start.x] {
			ctext = append(ctext, "-"+s)
			count.x++
		}
		for _, s := range y[done.y:start.y] {
			ctext = append(ctext, "+"+s)
			count.y++
		}

		const C = 3 // number of context lines
		if (end.x < len(x) || end.y < len(y)) &&
			(end.x-start.x < C || (len(ctext) > 0 && end.x-start.x < 2*C)) {
			for _, s := range x[start.x:end.x] {
				ctext = append(ctext, " "+s)
				count.x++
				count.y++
			}
			done = end
			continue
		}

		if len(ctext) > 0 {
			n := end.x - start.x
			if n > C {
				n = C
			}
			for _, s := range x[start.x : start.x+n] {
				ctext = append(ctext, " "+s)
				count.x++
				count.y++
			}
			done = pair{start.x + n, start.y + n}

			if count.x > 0 {
				chunk.x++
			}
			if count.y > 0 {
				chunk.y++
			}
			fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", chunk.x, count.x, chunk.y, count.y)
			for _, s := range ctext {
				out.WriteString(s)
			}
			count.x = 0
			count.y = 0
			ctext = ctext[:0]
		}

		if end.x >= len(x) && end.y >= len(y) {
			break
		}

		chunk = pair{end.x - C, end.y - C}
		for _, s := range x[chunk.x:end.x] {
			ctext = append(ctext, " "+s)
			count.x++
			count.y++
		}
		done = end
	}

	return out.Bytes()
}

func lines(x []byte) []string {
	l := strings.SplitAfter(string(x), "\n")
	if l[len(l)-1] == "" {
		l = l[:len(l)-1]
	} else {

		l[len(l)-1] += "\n\\ No newline at end of file\n"
	}
	return l
}

func tgs(x, y []string) []pair {

	m := make(map[string]int)
	for _, s := range x {
		if c := m[s]; c > -2 {
			m[s] = c - 1
		}
	}
	for _, s := range y {
		if c := m[s]; c > -8 {
			m[s] = c - 4
		}
	}

	var xi, yi, inv []int
	for i, s := range y {
		if m[s] == -1+-4 {
			m[s] = len(yi)
			yi = append(yi, i)
		}
	}
	for i, s := range x {
		if j, ok := m[s]; ok && j >= 0 {
			xi = append(xi, i)
			inv = append(inv, j)
		}
	}

	J := inv
	n := len(xi)
	T := make([]int, n)
	L := make([]int, n)
	for i := range T {
		T[i] = n + 1
	}
	for i := 0; i < n; i++ {
		k := sort.Search(n, func(k int) bool {
			return T[k] >= J[i]
		})
		T[k] = J[i]
		L[i] = k + 1
	}
	k := 0
	for _, v := range L {
		if k < v {
			k = v
		}
	}
	seq := make([]pair, 2+k)
	seq[1+k] = pair{len(x), len(y)} // sentinel at end
	lastj := n
	for i := n - 1; i >= 0; i-- {
		if L[i] == k && J[i] < lastj {
			seq[k] = pair{xi[i], yi[J[i]]}
			k--
		}
	}
	seq[0] = pair{0, 0} // sentinel at start
	return seq
}
