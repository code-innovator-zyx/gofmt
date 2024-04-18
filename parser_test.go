package main

import (
	"fmt"
	"testing"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/4/17 13:42
* @Package:
 */

func Test_innerStruct(t *testing.T) {

	var element = struct {
	}{}
	fmt.Println(sizeOf(element))
}

func Test_parseStruct(t *testing.T) {
	type args struct {
		structData []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{
			name: "struct",
			args: args{[]byte(`type People struct {
	b struct {
	}
	c     struct {
		a string
		c map[string]int
		b int32
	}
	Loves []int // 24
	Where []int
	e     []int
	MachineTime time.Time // 机审时间
	Name  string // 16
	Age   int    // 8
	a     int8
	has bool
	class Class
}                   //24   24 24`)}, want: []byte(``),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, count := parseStruct(tt.args.structData)
			t.Log(string(got))
			t.Log(count)
		})
	}
}
