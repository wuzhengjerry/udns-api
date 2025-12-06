package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	config "github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/frame/konghmac"
)

const (
	SendMDApiUri   = "/pigeon/v1/wechat_work/cs/markdown"
	SendMailApiUri = "/pigeon/v1/mail"
)

// FlyPigeonMDReqBody 请求体
type FlyPigeonMDReqBody struct {
	SendTo     []string `json:"sendTo"`
	MsgContent string   `json:"msgContent"`
	CorpID     string   `json:"corpId"`
	CorpSecret string   `json:"corpSecret"`
}

// FlyPigeonMailReqBody 请求体
type FlyPigeonMailReqBody struct {
	AppKey     string   `json:"appKey"`
	SysId      string   `json:"sysId"`
	SendFrom   string   `json:"sendFrom"`
	SendTo     []string `json:"sendTo"`
	SendCopy   []string `json:"sendCopy"`
	SendTitle  string   `json:"sendTitle"`
	MsgContent string   `json:"msgContent"`
	MailType   int      `json:"mailType"`
}

// SendMarkdownMessage 发送markdown格式的成本总览消息，到指定人员
func SendMarkdownMessage(msg, users string) error {

	userList := strings.Split(users, ";")
	reqBody := FlyPigeonMDReqBody{
		SendTo:     userList,
		MsgContent: msg,
		CorpID:     config.C().FlyPigeon.CorpID,
		CorpSecret: config.C().FlyPigeon.CorpSecret,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	url := config.C().FlyPigeon.Addr + SendMDApiUri

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	header := konghmac.GetAuthHeader(config.C().FlyPigeon.HMacID, config.C().FlyPigeon.HMacSecret, b)
	header["Content-Type"] = "application/json"

	//设置头部
	for k, v := range header {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("get business from fly pigeon error: %s", string(content))
		fmt.Println(errStr)
		return errors.New(errStr)
	}
	return nil
}

// SendMail 发送邮件消息
func SendMail(msg, users, subject string) error {

	userList := strings.Split(users, ";")
	reqBody := FlyPigeonMailReqBody{
		AppKey:     config.C().FlyPigeon.AppKey,
		SysId:      config.C().FlyPigeon.SysID,
		SendFrom:   config.C().FlyPigeon.MailSendFrom,
		SendTo:     userList,
		SendCopy:   []string{},
		SendTitle:  subject,
		MsgContent: msg,
		MailType:   1,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	url := config.C().FlyPigeon.Addr + SendMailApiUri

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	header := konghmac.GetAuthHeader(config.C().FlyPigeon.HMacID, config.C().FlyPigeon.HMacSecret, b)
	header["Content-Type"] = "application/json"

	//设置头部
	for k, v := range header {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("get business from fly pigeon error: %s", string(content))
		fmt.Println(errStr)
		return errors.New(errStr)
	}
	return nil
}
