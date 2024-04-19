package main

import (
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/4/19 17:33
* @Package:
 */

type People struct {
	b struct{} // 0
	c struct {
		a string
		c map[string]int
		b int32 // } haa struct {
	} // 8+8+4 =20
	Loves       []int         // 24
	MachineTime time.Time     // 24
	d           []int         //24
	Where       []int         //24
	e           []int         //24
	Name        string        // 16
	donot       interface{}   //16
	name        string        //16
	Age         int           // 8
	age         int           // 8
	inte        uintptr       //8
	sign        chan struct{} //8
	has         bool          //1
	a           int8          //1
}

func main() {
	//fmt.Println("before sort", unsafe.Sizeof(People{}))  // 256
}
