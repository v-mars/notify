package wechat

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/types"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	defaultMsgType  = "text"
	MsgTypeText     = "text"
	MsgTypeTextCard = "textcard"
	MsgTypeMarkdown = "markdown"
)

//"textcard": map[string]interface{}{
//"title":       title,
//"description": description,
//"url":         url,
//"btntext":     btntxt,
//},

// Err 微信返回错误
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// AccessToken 微信企业号请求Token
type accessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Err
	ExpiresInTime time.Time
}

// Wecom 微信企业号应用配置信息
type Wecom struct {
	types.WecomConfig
	TextCard       map[string]interface{} `json:"textcard"`
	Token          accessToken
	toParty, toTag []string
}

// Result 发送消息返回结果
type Result struct {
	Err
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"infvalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// Content 文本消息内容
type Content struct {
	Content string `json:"content"`
}

// Message 消息主体参数
type Message struct {
	ToUser   string                 `json:"touser"`
	ToParty  string                 `json:"toparty"`
	ToTag    string                 `json:"totag"`
	MsgType  string                 `json:"msgtype"`
	AgentID  int                    `json:"agentid"`
	Text     Content                `json:"text"`
	Markdown Content                `json:"markdown"`
	TextCard map[string]interface{} `json:"textcard"`
}

// NewWeChat init wechat notidy
func NewWeChat(cropID string, agentID int, agentSecret string,
	msgTextCard map[string]interface{}, toParty, toTag []string) *Wecom {
	defaultMsgType = MsgTypeText
	return &Wecom{
		WecomConfig: types.WecomConfig{
			CorpID:  cropID,
			AgentID: agentID,
			Secret:  agentSecret,
		},
		TextCard: msgTextCard,
		toParty:  toParty,
		toTag:    toTag,
	}
}

// Send format send msg to Message
func (c *Wecom) Send(tos []string, title, content string) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeWecom,
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
	msg := Message{
		ToUser:  strings.Join(tos, "|"),
		ToParty: strings.Join(c.toParty, "|"),
		ToTag:   strings.Join(c.toTag, "|"),
		MsgType: defaultMsgType,
		Markdown: Content{
			Content: title + "\n" + content,
		},
		TextCard: c.TextCard,
		Text: Content{
			Content: title + "\n" + content,
		},
		AgentID: c.AgentID,
	}
	if err = c.send(msg); err != nil {
		return sendResult, err

	}
	return sendResult, nil
}
func (c *Wecom) SendV2(tos, toParty, toTag []string, title, content string, msgTextCard map[string]interface{}) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeWecom,
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
		sendResult.Error = result.PtrOf(err.Error())
	}()
	msg := Message{
		ToUser:  strings.Join(tos, "|"),
		ToParty: strings.Join(toParty, "|"),
		ToTag:   strings.Join(toTag, "|"),
		MsgType: defaultMsgType,
		Markdown: Content{
			Content: title + "\n" + content,
		},
		TextCard: msgTextCard,
		Text: Content{
			Content: title + "\n" + content,
		},
		AgentID: c.AgentID,
	}
	if err = c.send(msg); err != nil {
		return sendResult, err

	}
	return sendResult, nil
}

// Send 发送信息
func (c *Wecom) send(msg Message) error {
	c.generateAccessToken()

	url := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + c.Token.AccessToken
	resultByte, err := notify.JSONPost(http.MethodPost, url, msg, http.DefaultClient, nil)
	if err != nil {
		err = errors.New("请求微信接口失败: " + err.Error())
		return err
	}
	r := Result{}
	err = json.Unmarshal(resultByte, &r)
	if err != nil {
		err = errors.New("解析微信接口返回数据失败: " + err.Error())
		return err
	}

	if r.ErrCode != 0 {
		err = errors.New("发送消息失败: " + r.ErrMsg)
		return err

	}

	if r.InvalidUser != "" || r.InvalidTag != "" || r.InvalidParty != "" {
		err = fmt.Errorf("消息发送成功, 但是有部分目标无法送达: %s %s %s", r.InvalidUser, r.InvalidParty, r.InvalidTag)
		return err
	}
	return nil
}

func (c *Wecom) SetMsgType(msgType string) {
	defaultMsgType = msgType
}

// generateAccessToken 生成会话token
func (c *Wecom) generateAccessToken() {
	var err error
	if c.Token.AccessToken == "" || c.Token.ExpiresInTime.Before(time.Now()) {
		c.Token, err = getAccessTokenFromWeixin(c.CorpID, c.Secret)
		if err != nil {
			return
		}
		c.Token.ExpiresInTime = time.Now().Add(time.Duration(c.Token.ExpiresIn-1000) * time.Second)
	}
}

const NotifyTypeWecom = "wecom"

func (c *Wecom) ChannelType() string {
	return NotifyTypeWecom
}

// 从微信服务器获取token
func getAccessTokenFromWeixin(cropID, secret string) (TokenSession accessToken, err error) {
	WxAccessTokenURL := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + cropID + "&corpsecret=" + secret

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	result, err := client.Get(WxAccessTokenURL)
	if err != nil {
		return
	}

	res, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return
	}

	defer result.Body.Close()

	err = json.Unmarshal(res, &TokenSession)
	if err != nil {
		return
	}

	if TokenSession.ExpiresIn == 0 || TokenSession.AccessToken == "" {
		err = fmt.Errorf("获取微信错误代码: %v, 错误信息: %v", TokenSession.ErrCode, TokenSession.ErrMsg)
		return
	}

	return
}
