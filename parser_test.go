package main

import (
	"bufio"
	"reflect"
	"testing"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2024/4/17 13:42
* @Package:
 */

func Test_innerStruct(t *testing.T) {
	type args struct {
		scanner *bufio.Scanner
		res     []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, count := innerStruct(tt.args.scanner, tt.args.res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("innerStruct() = %v, want %v", got, tt.want)
				t.Log(count)
			}
		})
	}
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

	Loves []int // 24
	Where []int
	e     []int
	Name  string // 16
	Age   int    // 8
	has   bool
	c struct {
		a string
		c map[string]int
		b int32
	}
	a     int8
}`)}, want: []byte(``),
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
