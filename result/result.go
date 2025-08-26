package result

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// SendResult 单个渠道的消息发送结果
type SendResult struct {
	ChannelType  string                    `json:"channel_type"`   // 渠道类型（如"email"、"sms"、"wecom"）
	Success      bool                      `json:"success"`        // 发送是否成功
	MessageID    string                    `json:"message_id"`     // 消息在系统中的唯一ID（关联主消息记录）
	ChannelMsgID *string                   `json:"channel_msg_id"` // 渠道返回的消息ID（如短信平台的msgid，可选）
	Error        *string                   `json:"error"`          // 失败原因（成功时为nil）
	SendTime     time.Time                 `json:"send_time"`      // 实际发送时间
	CostMs       int64                     `json:"cost_ms"`        // 发送耗时（毫秒）
	Cb           func(s *SendResult) error `json:"-"`              // 发送完成回调
}

type SendResults []*SendResult

func (s *SendResults) StatisticalResult() (success, failed int, err error) {
	var errs []string
	for _, v := range *s {
		if v.Success {
			success++
		} else {
			failed++
			if v.Error != nil {
				errs = append(errs, fmt.Sprintf("%s: %s", v.ChannelType, *v.Error))
			}
		}
	}
	return success, failed, errors.New(strings.Join(errs, ", "))
}
func (s *SendResults) ResultMsg() string {
	success, failed, err := s.StatisticalResult()
	if err != nil {
		return fmt.Sprintf("渠道发送成功: %d, 渠道发送失败: %d, 错误: %s", success, failed, err.Error())
	}
	return fmt.Sprintf("渠道发送成功: %d, 渠道发送失败: %d", success, failed)
}

func PtrOf[T any](v T) *T {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}
	return &v
}
