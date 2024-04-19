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
	src := "\t// ExtKeywordsResults KeyWords `json:\"-\" description:\"自定义关键词\"`}"
	fmt.Println(removeCommentString(src))
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
			args: args{[]byte("type ImageAbilityDetail struct {\n\tLevel1Code     string    `json:\"level1_code\"`            //违规的一级code\n\tLevel2Code     string    `json:\"level2_code,omitempty\"`  //违规的二级code\n\tLevel1Tag      string    `json:\"level1_tag\"`             //一级标签\n\tLevel2Tag      string    `json:\"level2_tag,omitempty\"`   //二级标签\n\tSuggestion     int       `json:\"suggestion\"`             //审核建议：1建议通过，2建议复审，3建议拦截\n\tLevel1Score    float64   `json:\"level1_score,omitempty\"` //一级标签置信度\n\tLevel2Score    float64   `json:\"level2_score,omitempty\"` //二级标签置信度\n\tFraction       float64   `json:\"fraction,omitempty\"`     //违规的分数\n\tKeywordsResult *KeyWords `json:\"-\" description:\"自定义关键词\"`\n\t// ExtKeywordsResults KeyWords `json:\"-\" description:\"自定义关键词\"`}}\n}")}, want: []byte(``),
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
