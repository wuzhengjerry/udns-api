package tool

import (
	"errors"
	"strings"

	"github.com/go-gomail/gomail"
	config "github.com/wuzhengjerry/udns-api/conf"
)

// SendEmail body支持html格式字符串
func SendEmail(users, subject, body string) (err error) {
	if users == "" {
		return errors.New("send user is nil")
	}
	c := config.C().Email
	sendTo := []string{}

	m := gomail.NewMessage()

	for _, tmp := range strings.Split(users, ";") {
		sendTo = append(sendTo, strings.TrimSpace(tmp))
	}
	m.SetHeader("To", sendTo...)
	//抄送列表，这里不涉及抄送场景，注释
	//if len(ep.CCers) != 0 {
	//	for _, tmp := range strings.Split(ep.CCers, ",") {
	//		userList = append(userList, strings.TrimSpace(tmp))
	//	}
	//	m.SetHeader("Cc", userList...)
	//}

	m.SetAddressHeader("From", c.FromEmail, "")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(c.ServerHost, c.ServerPort, c.FromEmail, c.Passwd)
	// 发送
	err = d.DialAndSend(m)
	if err != nil {
		return err
	}
	return err
}
