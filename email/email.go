package email

import (
	"crypto/tls"
	"github.com/v-mars/notify"
	"github.com/v-mars/notify/result"
	"time"

	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTP is email conf
type SMTP struct {
	SMTPHost   string
	Port       int
	Username   string
	Password   string
	From       string
	TLS        bool
	Anonymous  bool
	SkipVerify bool
}

// NewSMTP return a tls Smtp
func NewSMTP(smtphost string, port int, username, password, from string, tls, anonymous, skipVerify bool) notify.Sender {
	return &SMTP{
		SMTPHost:   smtphost,
		Username:   username,
		Password:   password,
		From:       from,
		TLS:        tls,
		Port:       port,
		Anonymous:  anonymous,
		SkipVerify: skipVerify,
	}
}

// Send send email to user
func (s *SMTP) Send(tos []string, title, content string) (sendResult *result.SendResult, err error) {
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
		sendResult.Error = result.PtrOf(err.Error())
	}()
	if s.SMTPHost == "" {
		return sendResult, fmt.Errorf("address is necessary")
	}
	var safeTos []string
	for _, to := range tos {
		err = CheckEmail(to)
		if err != nil {
			//log.Error("email check error", zap.Error(err))
			continue
		}
		safeTos = append(safeTos, to)
	}

	toAddr := strings.Join(safeTos, ";")

	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	header := make(map[string]string)
	header["From"] = s.From
	header["To"] = toAddr
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(title)))
	header["MIME-Version"] = "1.0"

	header["Content-TypeV1"] = "text/plain"
	header["Content-Transfer-Encoding"] = "base64"
	//header.Attach("./Dockerfile")   //添加附件

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(content))

	var auth smtp.Auth = nil
	if !s.Anonymous {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)
	}
	return sendResult, s.sendMail(auth, safeTos, []byte(message))
}

// sendMail will send mail to user
func (s *SMTP) sendMail(auth smtp.Auth, to []string, msg []byte) (err error) {
	if err := validateLine(s.From); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	var client *smtp.Client
	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.Port)
	if s.TLS {
		// httpsProxyURI, _ := url.Parse("https://your https proxy:443")
		// httpsDialer, err := proxy.FromURL(httpsProxyURI, HttpsDialer)

		tlsconfig := &tls.Config{
			InsecureSkipVerify: s.SkipVerify,
			ServerName:         s.SMTPHost,
		}
		var c *tls.Conn
		c, err = tls.Dial("tcp", addr, tlsconfig)

		if err != nil {
			return err
		}

		// tls.DialWithDialer(dialer *net.Dialer, network string, addr string, config *tls.NotifyConfig)
		client, err = smtp.NewClient(c, s.SMTPHost)
		if err != nil {
			return err
		}

		defer client.Close()
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return err
		}

		defer client.Close()

		if ok, _ := client.Extension("STARTTLS"); ok {
			config := &tls.Config{
				InsecureSkipVerify: s.SkipVerify,
				ServerName:         s.SMTPHost,
			}
			if err = client.StartTLS(config); err != nil {
				return err
			}
		}
	}
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}
	if err = client.Mail(s.From); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return client.Quit()

}

const NotifyTypeEmail = "email"

func (s *SMTP) ChannelType() string {
	return NotifyTypeEmail
}

func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return fmt.Errorf("smtp: A line must not contain CR or LF")
	}
	return nil
}
