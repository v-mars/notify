package email

import (
	"crypto/tls"
	"fmt"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/types"
	"gopkg.in/gomail.v2"
	"log"
	"time"
)

// MailboxConf 邮箱配置
type MailboxConf struct {
	// 邮件标题
	Title string `json:"title"`
	// 邮件内容
	Body string `json:"body"`
	// 邮件附件
	Attach []string `json:"attach"`
	// 收件人列表
	RecipientList []string `json:"recipient_list"`
	AttachList    []string `json:"attach_list"`

	types.EmailConfig
}

func NewMail(username, password, SMTPAddr, from string, SMTPPort int, tls bool, args ...any) *MailboxConf {
	im := MailboxConf{
		EmailConfig: types.EmailConfig{
			Username:   username,
			Password:   password,
			SMTPServer: SMTPAddr,
			SMTPPort:   SMTPPort,
			From:       from,
			TLS:        tls,
		},
	}
	if len(args) > 0 {
		im.AttachList = args[0].([]string)
	}
	return &im
}

func (mailConf *MailboxConf) Send(RecipientList []string, title, body string) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeEmail,
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

	m := gomail.NewMessage()
	m.SetHeader(`From`, mailConf.Username)
	//m.SetHeader(`To`, RecipientList...)
	m.SetHeader(`Subject`, title)
	m.SetBody(`text/html`, body)
	//m.Attach("./Dockerfile") //添加附件
	if len(mailConf.AttachList) > 0 {
		for _, v := range mailConf.AttachList {
			m.Attach(v)
		}
	}

	if len(RecipientList) == 0 {
		err = fmt.Errorf("发送邮件失败，邮箱接收者不能为空\n")
		log.Println(err)
		return sendResult, err
	}
	dialer := gomail.NewDialer(mailConf.SMTPServer, mailConf.SMTPPort, mailConf.Username, mailConf.Password)
	//dialer.TLSConfig = &tls.NotifyConfig{InsecureSkipVerify: true}
	//if mailConf.Tls{dialer.TLSConfig = &tls.NotifyConfig{InsecureSkipVerify: false}}

	// 逐个发送邮件，确保一个失败不会影响其他邮件的发送
	var failedRecipients []string
	var successRecipients []string
	var lastError error
	for _, recipient := range RecipientList {
		m.SetHeader(`To`, recipient)
		err = dialer.DialAndSend(m)
		if err != nil {
			failedRecipients = append(failedRecipients, recipient)
			lastError = fmt.Errorf("发送邮件到 %s 失败: %s", recipient, err.Error())
			log.Println(lastError)
		} else {
			successRecipients = append(successRecipients, recipient)
			log.Printf("发送邮件到 %s 成功", recipient)
		}
	}
	// 记录发送结果
	if len(failedRecipients) > 0 {
		log.Printf("部分邮件发送失败: %v", failedRecipients)
		if len(successRecipients) > 0 {
			log.Printf("部分邮件发送成功: %v", successRecipients)
		}
		// 如果有成功的邮件，不返回错误，但记录警告
		if len(successRecipients) > 0 {
			log.Printf("虽然部分邮件发送失败，但其他邮件发送成功")
			// 不返回错误，让调用方知道部分成功
			return sendResult, nil
		}
		// 如果所有邮件都失败，返回最后一个错误
		return sendResult, fmt.Errorf("所有邮件发送失败，最后的错误: %s", lastError.Error())
	} else {
		log.Println("所有邮件发送成功")
		return sendResult, nil
	}

	//log.Println("Send Email Success")
	//return sendResult, nil
}

func (mailConf *MailboxConf) ChannelType() string {
	return NotifyTypeEmail
}

func SendMailTest() {
	var mailConf MailboxConf
	mailConf.Title = "测试用gomail发送邮件"
	mailConf.Body = "Good Good Study, Day Day Up!!!!!!"
	mailConf.RecipientList = []string{`xxx@qq.com`}
	mailConf.Username = `xx@xx.cn`
	mailConf.Password = "xxx"
	mailConf.SMTPServer = `smtp.xx.com`
	mailConf.SMTPPort = 25

	m := gomail.NewMessage()
	m.SetHeader(`From`, mailConf.Username)
	m.SetHeader(`To`, mailConf.RecipientList...)
	m.SetHeader(`Subject`, mailConf.Title)
	m.SetBody(`text/html`, mailConf.Body)
	//m.Attach("./Dockerfile") //添加附件
	//d.TLSConfig = &tls.NotifyConfig{InsecureSkipVerify: true}
	dialer := gomail.NewDialer(mailConf.SMTPServer, mailConf.SMTPPort, mailConf.Username, mailConf.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := dialer.DialAndSend(m)
	if err != nil {
		log.Fatalf("Send Email Fail, %s", err.Error())
		return
	}
	log.Printf("Send Email Success")
}

/*

-----------------------------------
120 Service ready in nnn minutes.
服务在NNN时间内可用
-----------------------------------
125 Data connection already open; transfer starting.
数据连接已经打开，开始传送数据.
-----------------------------------
150 File status okay; about to open data connection.
文件状态正确，正在打开数据连接.
-----------------------------------
200 Command okay.
命令执行正常结束.
-----------------------------------
202 Command not implemented, superfluous at this site.
命令未被执行，此站点不支持此命令.
-----------------------------------
211 System status, or system help reply.
系统状态或系统帮助信息回应.
-----------------------------------
212 Directory status.
目录状态信息.
-----------------------------------
213 File status.
文件状态信息.
-----------------------------------
214 Help message.On how to use the server or the meaning of a particular non-standard command. This reply is useful only to the human user. 帮助信息。关于如何使用本服务器或特殊的非标准命令。此回复只对人有用。
-----------------------------------
215 NAME system type. Where NAME is an official system name from the list in the Assigned Numbers document.
NAME系统类型。
-----------------------------------
220 Service ready for new user.
新连接的用户的服务已就绪
-----------------------------------
221 Service closing control connection.
控制连接关闭
-----------------------------------
225 Data connection open; no transfer in progress.
数据连接已打开，没有进行中的数据传送
-----------------------------------
226 Closing data connection. Requested file action successful (for example, file transfer or file abort).
正在关闭数据连接。请求文件动作成功结束（例如，文件传送或终止）
-----------------------------------
227 Entering Passive Mode (h1,h2,h3,h4,p1,p2).
进入被动模式
-----------------------------------
230 User logged in, proceed. Logged out if appropriate.
用户已登入。 如果不需要可以登出。
-----------------------------------
250 Requested file action okay, completed.
被请求文件操作成功完成
-----------------------------------
257 "PATHNAME" created.
路径已建立
-----------------------------------
331 User name okay, need password.
用户名存在，需要输入密码
-----------------------------------
332 Need account for login.
需要登陆的账户
-----------------------------------
350 Requested file action pending further information
对被请求文件的操作需要进一步更多的信息
-----------------------------------
421 Service not available, closing control connection.This may be a reply to any command if the service knows it must shut down.
服务不可用，控制连接关闭。这可能是对任何命令的回应，如果服务认为它必须关闭
-----------------------------------
425 Can’t open data connection.
打开数据连接失败
-----------------------------------
426 Connection closed; transfer aborted.
连接关闭，传送中止。
-----------------------------------
450 Requested file action not taken.
对被请求文件的操作未被执行
-----------------------------------
451 Requested action aborted. Local error in processing.
请求的操作中止。处理中发生本地错误。
-----------------------------------
452 Requested action not taken. Insufficient storage space in system.File unavailable (e.g., file busy).
请求的操作没有被执行。 系统存储空间不足。 文件不可用
-----------------------------------
500 Syntax error, command unrecognized. This may include errors such as command line too long.
语法错误，不可识别的命令。 这可能是命令行过长。
-----------------------------------
501 Syntax error in parameters or arguments.
参数错误导致的语法错误
-----------------------------------
502 Command not implemented.
命令未被执行
-----------------------------------
503 Bad sequence of commands.
命令的次序错误。
-----------------------------------
504 Command not implemented for that parameter.
由于参数错误，命令未被执行
-----------------------------------
530 Not logged in.
没有登录
-----------------------------------
532 Need account for storing files.
存储文件需要账户信息
-----------------------------------
550 Requested action not taken. File unavailable (e.g., file not found, no access).
请求操作未被执行，文件不可用。
-----------------------------------
551 Requested action aborted. Page type unknown.
请求操作中止，页面类型未知
-----------------------------------
552 Requested file action aborted. Exceeded storage allocation (for current directory or dataset).
对请求文件的操作中止。 超出存储分配
-----------------------------------
553 Requested action not taken. File name not allowed
请求操作未被执行。 文件名不允许
-----------------------------------
-----------------------------------
这种错误跟http协议类似，大致是：
2开头－－成功
3开头－－权限问题
4开头－－文件问题
5开头－－服务器问题
对偶们最有用的：
421：一般出现在连接数多，需稍后在接；
530：密码错误；
550：目录或文件已经被删除。
*/
