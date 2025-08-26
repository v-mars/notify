package notify

import (
	"bytes"
	"encoding/json"
	"github.com/v-mars/notify/result"
	"io"
	"net/http"
)

// Sender it send notify to user
type Sender interface {
	Send(to []string, title string, content string) (*result.SendResult, error)
	ChannelType() string // 返回渠道类型（如"email"、"sms"）
}

// alarm notify
// mail
// chat
// dingding
// slack
// telegram
// server jiang
// lark

// JSONPost Post req json data to url
func JSONPost(method, url string, data interface{}, client *http.Client, headers map[string]string) ([]byte, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		//log.Error("client.Do",err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}
