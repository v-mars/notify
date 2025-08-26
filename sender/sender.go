package sender

import (
	"fmt"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/dingding"
	"github.com/v-mars/notify/email"
	"github.com/v-mars/notify/lark"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/sms"
	"github.com/v-mars/notify/types"
	"github.com/v-mars/notify/webhook"
	"github.com/v-mars/notify/wechat"
	"log"
	"sync"
	"time"
)

func NewNotifySender(conf *types.NotifyConfig, maxConcurrency int) *Manager {
	return &Manager{
		Conf:           conf,
		MaxConcurrency: maxConcurrency,
	}
}

// Manager 通知管理器，负责管理通知配置和发送消息
type Manager struct {
	Conf           *types.NotifyConfig
	MsgType        string `json:"msg_type" yaml:"msg_type"`
	ToParty, ToTag []string
	MaxConcurrency int // 最大并发数，默认为0表示无限制
}

// SendOptions 发送选项
type SendOptions struct {
	Channels []string // 指定发送渠道，为空时使用默认渠道
}

type Msg struct {
	Title     string
	EmailBody string
	ImBody    string
}

// Send 发送通知消息到指定的渠道
// 参数:
//
//	to: 接收人列表
//	title: 消息标题
//	content: 消息内容
//	opts: 发送选项，包括指定的渠道列表
//
// 返回:
//
//	发送结果列表和可能的错误
func (m *Manager) Send(to types.NotifyToIds, msg Msg, opts SendOptions) (result.SendResults, error) {
	if m == nil {
		return nil, fmt.Errorf("notify manager is nil")
	}
	// 确定发送渠道
	channels := opts.Channels
	if len(channels) == 0 {
		// 使用默认渠道列表
		channels = m.Conf.Channels
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("没有指定发送渠道")
	}

	// 并行发送消息
	var results []*result.SendResult
	var wg sync.WaitGroup
	resultChan := make(chan *result.SendResult, len(channels))

	// 创建带缓冲的信号量控制并发数
	maxConcurrency := m.MaxConcurrency
	if maxConcurrency <= 0 {
		// 默认并发数为渠道数量，无限制并发
		maxConcurrency = len(channels)
	}
	semaphore := make(chan struct{}, maxConcurrency)

	// 为每个渠道启动一个goroutine发送消息
	for _, channel := range channels {
		wg.Add(1)
		go func(ch string) {
			defer wg.Done()

			// 获取信号量，控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // 释放信号量
			var r *result.SendResult
			if ch == email.NotifyTypeEmail {
				r = m.SendToChannel(ch, to.GetToTagList(ch), msg.Title, msg.EmailBody)
			} else {
				r = m.SendToChannel(ch, to.GetToTagList(ch), msg.Title, msg.ImBody)
			}
			resultChan <- r
		}(channel)
	}

	// 等待所有发送完成
	wg.Wait()
	close(resultChan)

	// 收集结果
	for r := range resultChan {
		results = append(results, r)
	}

	return results, nil
}

// SendToChannel 向指定渠道发送消息
// 参数:
//
//	channel: 渠道类型，如"email", "sms"等
//	to: 接收人列表
//	title: 消息标题
//	content: 消息内容
//
// 返回:
//
//	发送结果
func (m *Manager) SendToChannel(channel string, to []string, title, content string) *result.SendResult {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic:", err)
		}
	}()
	var tmp []string
	for _, t := range to {
		if len(t) > 0 {
			tmp = append(tmp, t)
		}
	}
	to = tmp
	if len(to) == 0 {
		return result.PtrOf(result.SendResult{
			Success: false,
			Error:   result.PtrOf(fmt.Sprintf("发送渠道[%s]没有指定接收人", channel)),
		})
	}

	startTime := time.Now()

	// 创建默认失败结果
	r := &result.SendResult{
		ChannelType: channel,
		Success:     false,
		MessageID:   fmt.Sprintf("%d", time.Now().UnixNano()),
		SendTime:    startTime,
		CostMs:      0,
	}
	var sender notify.Sender
	var err error
	if m == nil {
		r.Error = result.PtrOf(fmt.Errorf("notify manager is nil").Error())
		return r
	}

	// 根据渠道类型创建对应的发送器
	switch channel {
	case email.NotifyTypeEmail:
		if m.Conf.Email != nil {
			sender = email.NewMail(
				m.Conf.Email.Username,
				m.Conf.Email.Password,
				m.Conf.Email.SMTPServer,
				m.Conf.Email.Username,
				m.Conf.Email.SMTPPort,
				m.Conf.Email.TLS,
			)
		}
	case sms.NotifyTypeSms:
		if m.Conf.Sms != nil {
			sender = sms.NewSms(
				m.Conf.Sms.Host,
				m.Conf.Sms.AccessKeyId,
				m.Conf.Sms.AccessKeySecret,
				m.Conf.Sms.SignName,
				m.Conf.Sms.TemplateCode,
			)
			// 设置短信签名和模板
		}
	case dingding.NotifyTypeDingDing:
		if m.Conf.Ding != nil {
			sender = dingding.NewDing(
				m.Conf.Ding.WebhookUrl,
				dingding.Sign, // 默认使用签名安全模式
				m.Conf.Ding.Secret,
			)
			// 设置消息类型
			d := sender.(*dingding.Ding)
			d.MsgType = m.Conf.Ding.MsgType
		}
	case lark.NotifyTypeLark:
		if m.Conf.Lark != nil {
			sender = lark.NewLark(
				m.Conf.Lark.WebhookUrl,
				lark.Sign,
				m.Conf.Lark.Secret,
			)
			// 设置消息类型
			l := sender.(*lark.Lark)
			l.MsgType = m.Conf.Lark.MsgType
		}
	case wechat.NotifyTypeWecom:
		if m.Conf.Wecom != nil {
			sender = wechat.NewWeChat(
				m.Conf.Wecom.CorpID,
				m.Conf.Wecom.AgentID,
				m.Conf.Wecom.Secret,
				nil,
				nil, nil,
			)
		}
	case webhook.NotifyTypeWebhook:
		if m.Conf.Webhook != nil {
			sender = webhook.NewWebhook(
				m.Conf.Webhook.URL,
				m.Conf.Webhook.Timeout,
				m.Conf.Webhook.Headers,
			)
		}
	default:
		err = fmt.Errorf("不支持的通知渠道: %s", channel)
	}

	// 检查发送器是否创建成功
	if sender == nil {
		err = fmt.Errorf("渠道 %s 配置不存在或无效", channel)
	}

	// 如果创建发送器过程中出现错误，直接返回错误结果
	if err != nil {
		errorMsg := err.Error()
		r.Error = &errorMsg
		r.CostMs = time.Since(startTime).Milliseconds()
		return r
	}

	// 发送消息
	sendResult, err := sender.Send(to, title, content)
	if err != nil {
		// 发送失败，记录错误信息
		errorMsg := err.Error()
		r.Error = &errorMsg
		r.CostMs = time.Since(startTime).Milliseconds()
		return r
	}

	// 发送成功，更新耗时信息
	sendResult.CostMs = time.Since(startTime).Milliseconds()
	return sendResult
}
