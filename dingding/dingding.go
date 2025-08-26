package dingding

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
	"net/url"
	"strconv"
	"time"
)

// getsign generate a sign when secure level is needsign
func getsign(secret string, now string) string {
	signstr := now + "\n" + secret
	// HmacSHA256
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(signstr))
	hm := h.Sum(nil)
	// Base64 encode
	b := base64.StdEncoding.EncodeToString(hm)
	// urlEncode
	sign := url.QueryEscape(b)
	return sign
}

// Secrue dingding secrue setting
// pls reading https://ding-doc.dingtalk.com/doc#/serverapi2/qf2nxq
type Secrue int

const (
	// CustomKey Custom keywords
	CustomKey Secrue = iota + 1
	// Sign need sign up
	Sign
	// IPCdir IP addres
	IPCdir
)

// Ding dingding alarm conf
type Ding struct {
	types.DingDing
	sl     Secrue
	Result *result.SendResult
}

// Result post resp
type Result struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type text struct {
	Content string `json:"content"`
}

type at struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

// SendMsg post json data
type SendMsg struct {
	MsgType string `json:"msgtype"`
	Text    text   `json:"text"`
	At      at     `json:"at"`
}

// NewDing init a Dingding send conf
func NewDing(webhookurl string, sl Secrue, secret string) *Ding {
	d := Ding{
		DingDing: types.DingDing{
			MsgType:    "text",
			WebhookUrl: webhookurl,
			Secret:     secret,
		},
		sl: sl,
		Result: &result.SendResult{
			ChannelType:  NotifyTypeDingDing,
			ChannelMsgID: nil,
			Success:      false,
			MessageID:    "",
			SendTime:     time.Now(),
			Error:        nil,
			CostMs:       0,
		},
	}

	return &d
}

// Send to notify tos is phone number
func (d *Ding) Send(tos []string, title string, content string) (sendResult *result.SendResult, err error) {
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
	if d.sl == Sign && len(d.Secret) > 0 {
		now := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
		sign := getsign(d.Secret, now)
		reqUrl += fmt.Sprintf("&timestamp=%s&sign=%s", now, sign)
	}
	sendMsg := SendMsg{
		MsgType: "text",
		Text: text{
			Content: title + "\n" + content + "\n",
		},
		At: at{
			AtMobiles: tos,
			IsAtAll:   false,
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
	if res.ErrCode != 0 {
		return sendResult, fmt.Errorf("errmsg: %s errcode: %d", res.ErrMsg, res.ErrCode)
	}
	return sendResult, nil
}

const NotifyTypeDingDing = "dingding"

func (d *Ding) ChannelType() string {
	return NotifyTypeDingDing
}
