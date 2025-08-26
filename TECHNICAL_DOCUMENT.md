# Notify 项目技术文档

## 项目概述

Notify 是一个统一消息通知管理系统，旨在提供一套支持多种消息渠道（如 Email、SMS、企业微信、飞书、钉钉、Webhook 等）并可灵活扩展的消息通知解决方案。系统提供标准化 API 接口，支持多渠道同时发送，实现消息模板管理、发送状态追踪及渠道配置管理。

## 核心功能

### 1. 渠道管理
- 支持动态添加/删除消息渠道（Email、SMS、企业微信、飞书、钉钉、Webhook 等）
- 每种渠道配置专属参数（如 Email 的 SMTP 服务器/端口/账号，SMS 的 API 密钥等）
- 支持渠道启用/禁用状态管理
- 通过接口标准化实现新渠道的无缝接入

### 2. 消息发送与调度
- 支持单次请求指定多个渠道同时发送
- 各渠道发送过程并行处理，提升发送效率
- 支持直接传入内容或指定模板 ID + 变量参数
- 支持设置消息优先级和定时发送

### 3. 发送状态跟踪
- 记录消息整体状态和各渠道单独状态
- 支持失败重试机制
- 详细记录每条消息的发送时间、耗时、响应数据、错误信息等

## 技术架构

### 1. 架构设计
采用分层架构设计：
- **接口层**：提供 RESTful API 供外部调用
- **业务层**：处理消息组装、渠道选择、发送逻辑
- **渠道层**：各消息渠道的具体实现（基于统一接口）
- **存储层**：消息、模板、渠道配置的数据持久化

### 2. 核心组件

#### 2.1 Sender 接口
所有消息渠道都必须实现 [Sender](file:///Users/zhangdonghai/work/dev/go/notify/notify.go#L12-L15) 接口：

```go
type Sender interface {
    Send(to []string, title string, content string) (*result.SendResult, error)
    ChannelType() string // 返回渠道类型（如"email"、"sms"）
}
```

#### 2.2 Manager 管理器
[Manager](file:///Users/zhangdonghai/work/dev/go/notify/sender/sender.go#L34-L39) 是通知管理器，负责管理通知配置和发送消息：

```go
type Manager struct {
    Conf           *types.NotifyConfig
    MsgType        string `json:"msg_type" yaml:"msg_type"`
    ToParty, ToTag []string
    MaxConcurrency int // 最大并发数，默认为0表示无限制
}
```

#### 2.3 消息发送结果
[SendResult](file:///Users/zhangdonghai/work/dev/go/notify/result/result.go#L12-L21) 表示单个渠道的消息发送结果：

```go
type SendResult struct {
    ChannelType  string     `json:"channel_type"`   // 渠道类型
    Success      bool       `json:"success"`        // 发送是否成功
    MessageID    string     `json:"message_id"`     // 消息在系统中的唯一ID
    ChannelMsgID *string    `json:"channel_msg_id"` // 渠道返回的消息ID
    Error        *string    `json:"error"`          // 失败原因
    SendTime     time.Time  `json:"send_time"`      // 实际发送时间
    CostMs       int64      `json:"cost_ms"`        // 发送耗时（毫秒）
}
```

## 支持的消息渠道

### 1. Email
通过 SMTP 协议发送邮件，支持 TLS 加密。

关键配置：
- SMTP 服务器地址
- SMTP 端口
- 发件人账号和密码（或授权码）
- TLS 加密选项

### 2. 钉钉 (DingDing)
通过钉钉机器人 Webhook 发送消息，支持签名安全模式。

关键配置：
- Webhook URL
- 安全签名密钥
- 消息类型

### 3. 飞书 (Lark)
通过飞书机器人 Webhook 发送消息，支持签名安全模式。

关键配置：
- Webhook URL
- 安全签名密钥
- 消息类型

### 4. 企业微信 (WeCom)
通过企业微信应用发送消息。

关键配置：
- 企业 ID (CorpID)
- 应用 AgentID
- 应用 Secret

### 5. 短信 (SMS)
通过短信服务提供商发送短信。

关键配置：
- 服务提供商 Host
- AccessKey ID 和 Secret
- 短信签名和模板

### 6. Webhook
通过 HTTP 请求发送消息到指定 URL。

关键配置：
- URL 地址
- 超时时间
- 自定义请求头

## 使用示例

### 1. 初始化 Manager
```go
config := &types.NotifyConfig{
    Channels: []string{"email", "dingding"},
    Email: &types.EmailConfig{
        SMTPServer: "smtp.example.com",
        SMTPPort:   587,
        Username:   "user@example.com",
        Password:   "password",
    },
    Ding: &types.DingDing{
        WebhookUrl: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
        Secret:     "your-secret",
    },
}

manager := sender.NewNotifySender(config, 10)
```

### 2. 发送消息
```go
to := types.NotifyToIds{
    {Email: "user@example.com", Ding: "123456789"},
}

msg := sender.Msg{
    Title:     "通知标题",
    EmailBody: "邮件内容",
    ImBody:    "IM消息内容",
}

results, err := manager.Send(to, msg, sender.SendOptions{})
if err != nil {
    // 处理错误
}

// 处理发送结果
for _, result := range results {
    if result.Success {
        fmt.Printf("渠道 %s 发送成功\n", result.ChannelType)
    } else {
        fmt.Printf("渠道 %s 发送失败: %s\n", result.ChannelType, *result.Error)
    }
}
```

## 扩展新渠道

要添加新的消息渠道，需要：

1. 实现 [Sender](file:///Users/zhangdonghai/work/dev/go/notify/notify.go#L12-L15) 接口
2. 在 [SendToChannel](file:///Users/zhangdonghai/work/dev/go/notify/sender/sender.go#L109-L257) 方法中添加渠道类型判断和初始化逻辑
3. 在 [NotifyConfig](file:///Users/zhangdonghai/work/dev/go/notify/types/types.go#L12-L21) 中添加相应的配置结构体

示例：
```go
type NewChannel struct {
    // 渠道特定配置
}

func (n *NewChannel) Send(to []string, title, content string) (*result.SendResult, error) {
    // 实现发送逻辑
}

func (n *NewChannel) ChannelType() string {
    return "newchannel"
}
```

## 性能与可靠性

### 性能要求
- 单节点支持每秒处理≥500 条消息请求
- 多渠道并发发送时，单条消息整体处理延迟≤1 秒

### 可靠性要求
- 消息不丢失：通过持久化确保消息至少被处理一次
- 失败重试：临时错误自动重试，重试策略可配置
- 系统可用性：≥99.9%

## 安全性

### 安全措施
- 敏感配置加密：渠道的密钥、账号等信息加密存储
- 消息内容加密：支持对敏感消息内容加密传输
- 接口鉴权：所有 API 接口需通过 Token 或签名验证