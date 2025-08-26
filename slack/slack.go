package slack

import (
	"fmt"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/result"
	"net/http"
	"time"
)

// Slack send conf
type Slack struct {
	webhookurl string
	httpclient *http.Client
}

// SendMsg post json data
type SendMsg struct {
	Text string `json:"text"`
	// Channels string `json:"channel"`
	Username string `json:"username"`
	// PreText string `json:"pretext"`
}

// NewSlack init
func NewSlack(webhook string) *Slack {
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	return &Slack{
		webhookurl: webhook,
		httpclient: client,
	}
}

// Send will send send msg to slack channel
func (s *Slack) Send(tos []string, title string, content string) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeSlack,
		ChannelMsgID: nil,
		Success:      false,
		MessageID:    "",
		SendTime:     time.Now(),
		Error:        nil,
		CostMs:       0,
	}
	defer func() {
		sendResult.CostMs = time.Now().Sub(sendResult.SendTime).Milliseconds()
		sendResult.ChannelMsgID = result.PtrOf(fmt.Sprintf("%d", time.Now().UnixNano()))
		sendResult.Success = err == nil
		sendResult.MessageID = *sendResult.ChannelMsgID
		if err != nil {
			sendResult.Error = result.PtrOf(err.Error())
		}
	}()
	sendmsg := SendMsg{
		Text: content + "\n" + content,
	}
	resp, err := notify.JSONPost(http.MethodPost, s.webhookurl, sendmsg, s.httpclient, nil)
	if err != nil {
		return sendResult, err
	}
	if string(resp) != "ok" {
		err = fmt.Errorf("send data to slack failed error:%s", resp)
		return sendResult, err
	}
	return sendResult, nil
}

const NotifyTypeSlack = "slack"

func (s *Slack) ChannelType() string {
	return NotifyTypeSlack
}
