package sms

import (
	"fmt"
	"testing"
)

func TestNewSms(t *testing.T) {
	api := NewSms("dysmsapi.aliyuncs.com", "xx",
		"xx", "xx", "xx")
	err := api.Send([]string{"1861651xxxx"}, "测试", `{"code":"123456"}`)
	fmt.Println(err)
}
