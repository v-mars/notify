package types

import (
	"time"
)

const KvNotifyName = "notify"

// NotifyConfig 通知系统配置结构体
type NotifyConfig struct {
	Channels []string     `json:"channels" yaml:"channels"`
	Email    *EmailConfig `json:"email" yaml:"email"`
	Lark     *Lark        `json:"lark" yaml:"lark"`
	Ding     *DingDing    `json:"dingding" yaml:"dingding"`
	Wecom    *WecomConfig `json:"wecom" yaml:"wecom"`
	Sms      *SmsConfig   `json:"sms" yaml:"sms"`
	Webhook  *Webhook     `json:"webhook" yaml:"webhook"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	// SMTP 服务器地址， QQ邮箱是smtp.qq.com
	SMTPServer string `json:"smtpServer" yaml:"smtpServer"`
	// SMTP端口 QQ邮箱是25
	SMTPPort int    `json:"smtpPort" yaml:"smtpPort"`
	From     string `json:"from" yaml:"from"`
	TLS      bool   `json:"tls" yaml:"tls"`
	// 发件人账号
	Username string `json:"username" yaml:"username"`
	// 发件人密码，QQ邮箱这里配置授权码
	Password string `json:"password" yaml:"password"`
}

// WecomConfig 企业微信配置
type WecomConfig struct {
	URL            string `json:"url" yaml:"url"`
	CorpID         string `json:"corp_id" yaml:"corp_id"`
	AgentID        int    `json:"agent_id" yaml:"agent_id"`
	Secret         string `json:"secret" yaml:"secret"`
	AddrBookSecret string `json:"addr_book_secret" yaml:"addr_book_secret"`
	Receivers      string `json:"receivers" yaml:"receivers"`
}

type Lark struct {
	MsgType    string `json:"msg_type" yaml:"msg_type"`
	WebhookUrl string `json:"webhook_url" yaml:"webhook_url"`
	Secret     string `json:"secret" yaml:"secret"`
}

type DingDing struct {
	MsgType    string `json:"msg_type" yaml:"msg_type"`
	WebhookUrl string `json:"webhook_url" yaml:"webhook_url"`
	Secret     string `json:"secret" yaml:"secret"`
}

// SmsConfig 短信配置
type SmsConfig struct {
	Host            string `json:"host" yaml:"host"`
	AccessKeyId     string `json:"access_key_id" yaml:"access_key_id"`         // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret" yaml:"access_key_secret"` // 访问密钥
	TemplateCode    string `json:"template_code" yaml:"template_code"`
	SignName        string `json:"sign_name" yaml:"sign_name"`
}

// Webhook represents a webhook notification configuration
type Webhook struct {
	URL     string            `json:"url" yaml:"url"`
	Timeout time.Duration     `json:"timeout" yaml:"timeout"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

type NotifyToId struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Wecom   string `json:"wecom"`
	Ding    string `json:"ding"`
	Lark    string `json:"lark"`
	Webhook string `json:"webhook"`
}

type NotifyToIds []NotifyToId

func (s *NotifyToIds) GetToTagList(sender string) []string {
	var tags []string
	for _, n := range *s {
		tag := ""
		switch sender {
		case "email":
			tag = n.Email
		case "sms":
			tag = n.Phone
		case "wecom":
			tag = n.Wecom
		case "lark":
			tag = n.Lark
		case "dingding":
			tag = n.Ding
		case "webhook":
			tag = n.Webhook
		}
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
