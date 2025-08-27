package email

import (
	"testing"
)

func TestNewSMTP(t *testing.T) {
	//

	mail := NewMail("xx@qq.com", "xx", "smtp.qq.com",
		"xx", 465, false)
	_, _ = mail.Send([]string{`xx@qq.com`, "xx@xx.cn"}, "测试用gomail发送邮件", "Good Good Study, Day Day Up!!!!!!")
}
