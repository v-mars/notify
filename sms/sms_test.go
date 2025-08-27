package sms

import (
	"fmt"
	"testing"
)

func TestNewSms(t *testing.T) {
	api := NewSms("dysmsapi.aliyuncs.com", "xx",
		"xx", "xx", "xx")
	err, _ := api.Send([]string{"186xxxxxxxx"}, "测试", `{"code":"123456"}`)
	fmt.Println(err)
}
