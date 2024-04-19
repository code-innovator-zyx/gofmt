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
	} // 0
	c struct {
		a string
		c map[string]int
		b int32 // } haa struct {
	}
	Loves       []int // 24
	Where       []int //24
	e           []int //24
	d           []int //24
	MachineTime time.Time
	Name        string // 16
	donot       interface{}
	name        string
	Age         int     // 8
	inte        uintptr //8
	age         int
	a           int8 //1
	has         bool //1
	//The following fields do not participate in byte alignment sorting. You can make adjustments by yourself
	class  Class
	class3 B
}

type Class struct {
	Where     []int
	Name      string
	HasPeople int
}

type A struct{ HasPeople int }

type B struct{ HasPeople int }

// ImageDebugResponse 图像Debug模式
type ImageDebugResponse struct {
	Data struct {
		RequestId    string      `json:"request_id"`              // 唯一请求ID
		CallbackInfo string      `json:"callback_info,omitempty"` // 回调信息
		Details      interface{} `json:"details"`
	} `json:"data,omitempty"`
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

type ImageAbilityDetail struct {
	Level2Code     string    `json:"level2_code,omitempty"`  //违规的二级code
	Level1Code     string    `json:"level1_code"`            //违规的一级code
	Level2Tag      string    `json:"level2_tag,omitempty"`   //二级标签
	Level1Tag      string    `json:"level1_tag"`             //一级标签
	Suggestion     int       `json:"suggestion"`             //审核建议：1建议通过，2建议复审，3建议拦截
	Level2Score    float64   `json:"level2_score,omitempty"` //二级标签置信度
	Level1Score    float64   `json:"level1_score,omitempty"` //一级标签置信度
	Fraction       float64   `json:"fraction,omitempty"`     //违规的分数
	KeywordsResult *KeyWords `json:"-" description:"自定义关键词"`
	// ExtKeywordsResults KeyWords `json:"-" description:"自定义关键词"`}}
}

type KeyWords struct {
	HitInfo     string            `json:"hit_info"`              // 文本违规涉及的关键字
	Location    string            `json:"location,omitempty"`    // 关键字所在的位置
	Tag         string            `json:"tag"`                   // 关键字命中的标签
	Description string            `json:"description,omitempty"` // 关键字描述
	TagL2       string            `json:"tagL2,omitempty"`
	TagL1       string            `json:"tagL1,omitempty"`
	Diff        int               `json:"diff,omitempty"` //
	a           map[string]string `json:"a"`
}
