package notify

import (
	"net/http"
	"testing"
)

func Test_JSONPost(t *testing.T) {
	_, err := JSONPost(http.MethodPost, "http://webhook.test", nil, http.DefaultClient, nil)
	if err != nil {
		t.Fatal(err)
	}
}

//// 消息发送器接口（所有渠道实现此接口）
//type MessageSender interface {
//	Send(ctx context.Context, message *Message) (*SendResult, error)
//	ChannelType() string // 返回渠道类型（如"email"、"sms"）
//	IsEnabled() bool     // 检查渠道是否启用
//}
//
//// 消息发送服务
//type NotificationService struct {
//	senderFactory *SenderFactory // 渠道工厂，用于创建具体发送器
//	templateStore TemplateStore  // 模板存储
//	messageStore  MessageStore   // 消息存储
//	retryStrategy RetryStrategy  // 重试策略
//}
//
//// 发送消息（支持多渠道）
//func (s *NotificationService) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
//	// 1. 解析请求（获取接收人、渠道列表、内容/模板）
//	// 2. 验证渠道状态
//	// 3. 组装消息（替换模板变量）
//	// 4. 并发调用各渠道发送器
//	// 5. 记录发送结果
//	// 6. 处理失败重试
//}
//
//// 消息渠道配置
//type ChannelConfig struct {
//	ID        string            `json:"id"`
//	Type      string            `json:"type"` // email, sms, wecom, etc.
//	Name      string            `json:"name"`
//	Config    map[string]string `json:"config"` // 渠道专属配置，如SMTP信息
//	Enabled   bool              `json:"enabled"`
//	CreatedAt time.Time         `json:"created_at"`
//	UpdatedAt time.Time         `json:"updated_at"`
//}
//
//// 消息模板
//type Template struct {
//	ID           string    `json:"id"`
//	Name         string    `json:"name"`
//	ChannelTypes []string  `json:"channel_types"` // 关联的渠道类型
//	Content      string    `json:"content"`       // 模板内容，含变量占位符
//	Subject      string    `json:"subject"`       // 仅用于邮件等有主题的渠道
//	Version      int       `json:"version"`
//	CreatedAt    time.Time `json:"created_at"`
//}
//
//// 消息主记录
//type Message struct {
//	ID           string            `json:"id"`
//	BusinessType string            `json:"business_type"` // 业务类型，如"verify_code"
//	Priority     string            `json:"priority"`      // high, medium, low
//	Status       string            `json:"status"`        // pending, sending, completed, failed
//	Recipients   []Recipient       `json:"recipients"`    // 接收人列表
//	TemplateID   *string           `json:"template_id"`   // 模板ID（可选）
//	Variables    map[string]string `json:"variables"`     // 模板变量（可选）
//	Content      *string           `json:"content"`       // 直接指定的内容（可选）
//	CreatedAt    time.Time         `json:"created_at"`
//	SentAt       *time.Time        `json:"sent_at"`
//}
//
//// 渠道发送记录
//type ChannelMessageRecord struct {
//	ID          string     `json:"id"`
//	MessageID   string     `json:"message_id"`
//	ChannelType string     `json:"channel_type"`
//	Status      string     `json:"status"`    // success, failed
//	ErrorMsg    *string    `json:"error_msg"` // 失败原因
//	SendTime    *time.Time `json:"send_time"`
//	CostMs      int64      `json:"cost_ms"` // 发送耗时（毫秒）
//}
