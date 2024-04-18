package main

/*
* @Author: zouyx
* @Email:
* @Date:   2024/4/16 14:39
* @Package:
 */

type data struct {
	score uint16 //置信度最高的第一个弹出
	res   []byte
}

type sortHeap []data

func (h sortHeap) Len() int {
	return len(h)
}

func (h sortHeap) Less(i, j int) bool {
	if h[i].score == h[j].score {
		return len(h[i].res) > len(h[j].res)
	}
	return h[i].score > h[j].score
}

func (h *sortHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *sortHeap) Push(x any) {
	*h = append(*h, x.(data))
}

// Pop 弹出切片内剩余的最大的一个元素   只能是接口类型
func (h *sortHeap) Pop() any {
	res := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return res
}
