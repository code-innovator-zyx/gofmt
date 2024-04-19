package main

import (
	"bytes"
	"strings"
)

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/18 18:35
* @Package:
 */

// RemoveComment filters comment content
func removeCommentString(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		line = line[:commentIdx]
	}
	tagIdx := strings.Index(line, "`")
	if tagIdx >= 0 {
		return strings.TrimSpace(line[:tagIdx])
	}
	return strings.TrimSpace(line)
}

func removeCommentByte(line []byte) []byte {
	commentIdx := bytes.Index(line, []byte("//"))
	if commentIdx >= 0 {
		return bytes.TrimSpace(line[:commentIdx])
	}
	return bytes.TrimSpace(line)
}
