package lark

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/types"
	"net/http"
	"time"
)

func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

type Secure uint

const (
	// CustomKey Custom keywords
	CustomKey Secure = iota + 1
	// Sign need sign up
	Sign
	// IPCdir IP addres
	IPCdir
)

// Lark alarm conf
type Lark struct {
	types.Lark
	Sl           Secure `json:"sl"`
	Result       *result.SendResult
	TitleTag     string
	Subtitle     string
	CardTemplate string // 卡片颜色主题：red, blue, green, orange 等
	ElementsTag  string
}

// Result post resp
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Content struct {
	Text string `json:"text"`
}

// SendMsg post json data
type SendMsg struct {
	Timestamp string         `json:"timestamp"`
	Sign      string         `json:"sign"`
	MsgType   string         `json:"msg_type"`
	Content   Content        `json:"content"`
	Card      map[string]any `json:"card"`
}

// NewLark init a Lark send conf
func NewLark(webHookUrl string, sl Secure, secret string) *Lark {
	d := Lark{
		Lark: types.Lark{
			MsgType:    "text",
			Secret:     secret,
			WebhookUrl: webHookUrl,
		},
		Sl: sl,
		Result: &result.SendResult{
			ChannelType:  NotifyTypeLark,
			ChannelMsgID: nil,
			Success:      false,
			MessageID:    "",
			SendTime:     time.Now(),
			Error:        nil,
			CostMs:       0,
		},
		TitleTag:     "plain_text",
		CardTemplate: "blue",
		ElementsTag:  "lark_md",
	}

	return &d
}

// Send to notify tos is phone number
func (d *Lark) Send(tos []string, title string, content string) (sendResult *result.SendResult, err error) {
	sendResult = d.Result
	defer func() {
		sendResult.CostMs = time.Now().Sub(sendResult.SendTime).Milliseconds()
		sendResult.ChannelMsgID = result.PtrOf(fmt.Sprintf("%d", time.Now().UnixNano()))
		sendResult.Success = err == nil
		sendResult.MessageID = *sendResult.ChannelMsgID
		if err != nil {
			sendResult.Error = result.PtrOf(err.Error())
		}
	}()
	var reqUrl = d.WebhookUrl
	timestamp := time.Now().Unix()
	sign := ""
	if len(d.Secret) > 0 {
		if sign, err = GenSign(d.Secret, timestamp); err != nil {
			return sendResult, err
		}
	}

	sendMsg := SendMsg{
		Timestamp: fmt.Sprintf("%d", timestamp),
		Sign:      sign,
		MsgType:   d.MsgType,
		Content: Content{
			Text: title + "\n" + content + "\n",
		},
		Card: map[string]any{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     d.TitleTag,
					"content": title, // 卡片标题
				},
				"subtitle": map[string]interface{}{
					"tag":     d.TitleTag,
					"content": d.Subtitle, // 卡片标题
				},
				"template": d.CardTemplate, // 卡片颜色主题：red, blue, green, orange 等
			},
			"elements": []interface{}{
				map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     d.ElementsTag, // 飞书支持的 Markdown 标签
						"content": content,
					},
				},
			},
		},
	}

	resp, err := notify.JSONPost(http.MethodPost, reqUrl, sendMsg, http.DefaultClient, nil)
	if err != nil {
		return sendResult, err
	}
	res := Result{}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return sendResult, err
	}
	if res.Code != 0 {
		return sendResult, fmt.Errorf("errmsg: %s errcode: %d", res.Msg, res.Code)
	}
	return sendResult, nil
}

const NotifyTypeLark = "lark"

func (d *Lark) ChannelType() string {
	return NotifyTypeLark
}
