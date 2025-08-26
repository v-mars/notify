package sender

import (
	"github.com/v-mars/notify/types"
	"reflect"
	"testing"
)

func TestNewNotifySender(t *testing.T) {
	// 创建测试用的NotifyConfig
	testConfig := &types.NotifyConfig{
		Channels: []string{"email", "sms"},
		Email: &types.EmailConfig{
			SMTPServer: "smtp.test.com",
			SMTPPort:   587,
			Username:   "test@test.com",
			Password:   "password",
		},
		Sms: &types.SmsConfig{
			Host:            "sms.test.com",
			AccessKeyId:     "test-key-id",
			AccessKeySecret: "test-key-secret",
			SignName:        "TestSign",
			TemplateCode:    "TestTemplate",
		},
	}

	tests := []struct {
		name            string
		conf            *types.NotifyConfig
		maxConcurrency  int
		expectedManager *Manager
	}{
		{
			name:           "Test with valid config and concurrency",
			conf:           testConfig,
			maxConcurrency: 5,
			expectedManager: &Manager{
				Conf:           testConfig,
				MaxConcurrency: 5,
			},
		},
		{
			name:           "Test with zero concurrency",
			conf:           testConfig,
			maxConcurrency: 0,
			expectedManager: &Manager{
				Conf:           testConfig,
				MaxConcurrency: 0,
			},
		},
		{
			name:           "Test with negative concurrency",
			conf:           testConfig,
			maxConcurrency: -1,
			expectedManager: &Manager{
				Conf:           testConfig,
				MaxConcurrency: -1,
			},
		},
		{
			name:           "Test with nil config",
			conf:           nil,
			maxConcurrency: 3,
			expectedManager: &Manager{
				Conf:           nil,
				MaxConcurrency: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNotifySender(tt.conf, tt.maxConcurrency)

			// 检查返回的Manager是否正确
			if !reflect.DeepEqual(got.Conf, tt.expectedManager.Conf) {
				t.Errorf("NewNotifySender().Conf = %v, want %v", got.Conf, tt.expectedManager.Conf)
			}

			if got.MaxConcurrency != tt.expectedManager.MaxConcurrency {
				t.Errorf("NewNotifySender().MaxConcurrency = %v, want %v", got.MaxConcurrency, tt.expectedManager.MaxConcurrency)
			}
		})
	}
}