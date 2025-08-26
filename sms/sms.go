package sms

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	log "github.com/sirupsen/logrus"
	"github.com/v-mars/notify/result"
	"github.com/v-mars/notify/types"
	"strings"
	"time"
)

func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

type Secure uint

const (
	// CustomKey Custom keywords
	CustomKey Secure = iota + 1
	// Sign need sign up
	Sign
	// IPCdir IP addres
	IPCdir
)

// SmsConf alarm conf
type SmsConf struct {
	Gw string `json:"gw"`
	types.SmsConfig
	SendMsg
}

// Result post resp
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Content struct {
	Text string `json:"text"`
}

// SendMsg post json data
type SendMsg struct {
	PhoneNumbers  string `json:"PhoneNumbers"`  // 支持向不同的手机号码发送短信，手机号码之间以半角逗号（,）分隔。上限为 1000 个手机号码。批量发送相对于单条发送，及时性稍有延迟。验证码类型的短信，建议单条发送。
	SignName      string `json:"SignName"`      // 短信签名名称
	TemplateCode  string `json:"TemplateCode"`  // 短信模板 Code。
	TemplateParam string `json:"TemplateParam"` // 短信模板变量对应的实际值。支持传入多个参数。{"name":"张三","number":"1390000****"}
}

// NewSms init a sms send conf
func NewSms(host, accessKeyId, accessKeySecret, SignName, templateCode string) *SmsConf {
	d := SmsConf{
		SmsConfig: types.SmsConfig{
			Host:            host,
			AccessKeyId:     accessKeyId,
			AccessKeySecret: accessKeySecret,
			SignName:        SignName,
			TemplateCode:    templateCode,
		},
		SendMsg: SendMsg{
			SignName:     SignName,
			TemplateCode: templateCode,
		},
	}

	return &d
}

// Send to notify tos is phone number content: {"name":"张三","number":"1390000****"}
func (d *SmsConf) Send(tos []string, title string, content string) (sendResult *result.SendResult, err error) {
	sendResult = &result.SendResult{
		ChannelType:  NotifyTypeSms,
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
	if d.Gw == "tencent" {
		return sendResult, d.TencentSender(tos, title, content)
	}
	return sendResult, d.AliYunSender(tos, title, content)
}

func (d *SmsConf) TencentSender(tos []string, title string, content string) (_err error) {
	client, _err := d.NewAliYunClient()
	if _err != nil {
		return _err
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      &d.SmsConfig.SignName,
		TemplateCode:  &d.SmsConfig.TemplateCode,
		PhoneNumbers:  tea.String(strings.Join(tos, ",")),
		TemplateParam: tea.String(content),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, _err = client.SendSmsWithOptions(sendSmsRequest, runtime)
		if _err != nil {
			return _err
		}

		return nil
	}()

	if tryErr != nil {
		var _er = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			_er = _t
		} else {
			_er.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		//fmt.Println(tea.StringValue(_er.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(_er.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			log.Info(recommend)
		}
		_, _err = util.AssertAsString(_er.Message)
		if _err != nil {
			return _err
		}
	}
	return nil
}

func (d *SmsConf) AliYunSender(tos []string, title string, content string) (_err error) {
	client, _err := d.NewAliYunClient()
	if _err != nil {
		return _err
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      &d.SmsConfig.SignName,
		TemplateCode:  &d.SmsConfig.TemplateCode,
		PhoneNumbers:  tea.String(strings.Join(tos, ",")),
		TemplateParam: tea.String(content),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, _err = client.SendSmsWithOptions(sendSmsRequest, runtime)
		if _err != nil {
			return _err
		}

		return nil
	}()

	if tryErr != nil {
		var _er = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			_er = _t
		} else {
			_er.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		//fmt.Println(tea.StringValue(_er.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(_er.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			log.Info(recommend)
		}
		_, _err = util.AssertAsString(_er.Message)
		if _err != nil {
			return _err
		}
	}
	return nil
}

// Description:
//
// 使用AK&SK初始化账号Client
//
// @return Client
//
// @throws Exception
func (d *SmsConf) NewAliYunClient() (_result *dysmsapi20170525.Client, _err error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: tea.String(d.AccessKeyId),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: tea.String(d.AccessKeySecret),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String(d.Host) // "dysmsapi.aliyuncs.com"
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

const NotifyTypeSms = "sms"

func (d *SmsConf) ChannelType() string {
	return NotifyTypeSms
}
