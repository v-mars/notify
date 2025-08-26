package wechat

import "testing"

func TestNewWeChat(t *testing.T) {
	api := NewWeChat("xx", 100, "xx",
		nil, nil, nil)
	_, err := api.Send([]string{"000"}, "测试", "测试")
	if err != nil {
		t.Error(err)
		return
	}
}
