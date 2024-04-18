package main

import (
	"container/heap"
	"fmt"
	"testing"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/4/17 17:06
* @Package:
 */

func Test_Heap(t *testing.T) {
	h := &sortHeap{
		// 初始化你的堆
	}
	// 初始化堆
	heap.Init(h)
	// 添加元素到堆中
	heap.Push(h, data{score: 1, res: []byte{1}})
	heap.Push(h, data{score: 3, res: []byte{3}})
	heap.Push(h, data{score: 3, res: []byte{4}})
	heap.Push(h, data{score: 3, res: []byte{5}})
	heap.Push(h, data{score: 3, res: []byte{6}})
	heap.Push(h, data{score: 3, res: []byte{7}})
	heap.Push(h, data{score: 2, res: []byte{2}})

	// 从堆中弹出元素
	for h.Len() > 0 {
		maxElement := heap.Pop(h).(data)
		fmt.Printf("Popped: %v\n", maxElement)
	}
}
