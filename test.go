package main

import "time"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/4/15 13:47
* @Package:
 */
type People struct {
	b struct {
	}
	c struct {
		a string
		c map[string]int
		b int32
	}
	MachineTime time.Time // 机审时间
	Loves       []int     // 24
	Where       []int
	e           []int
	d           []int
	Name        string // 16
	Age         int    // 8
	inte        uintptr
	a           int8
	has         bool
	//The following fields do not participate in byte alignment sorting. You can make adjustments by yourself
	class Class
}                   //24   24 24
type Class struct { //25  25
	Where     []int
	Name      string
	HasPeople int
}

type A struct{ HasPeople int } //30   30
type B struct{ HasPeople int } // 33   31
