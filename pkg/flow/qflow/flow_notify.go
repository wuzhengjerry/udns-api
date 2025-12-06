package qflow

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/tool"
)

// NotifyTask 发送任务通知
func (s *service) NotifyTask(taskID string) {
	t, err := s.DescribeTask(flow.SystemUserName, taskID, false)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"err_msg": err,
		}).Error("notify task error")
	}
	switch t.Status {
	case flow.DoingStatus:
		for _, i := range t.Steps {
			if i.Status == flow.DoingStatus {
				err = tool.SendMarkdownMessage(GenMsgContent(flow.DoingStatus, t.Name, taskID), i.Owner)
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"taskID": taskID,
						"error":  err,
					}).Error("send flyPigeon msg error:", err)
				}
			}
		}
	case flow.DoneStatus:
		err = tool.SendMarkdownMessage(GenMsgContent(flow.DoneStatus, t.Name, taskID), t.Creator)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"taskID": taskID,
				"error":  err,
			}).Error("send flyPigeon msg error:", err)
		}
	case flow.StopStatus:
		err = tool.SendMarkdownMessage(GenMsgContent(flow.StopStatus, t.Name, taskID), t.Creator)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"taskID": taskID,
				"error":  err,
			}).Error("send flyPigeon msg error:", err)
		}
	default:
		s.l.WithFields(logrus.Fields{
			"taskID": taskID,
			"error":  "bad notify status",
		}).Error("notify task status error")
	}

	return
}

// GenMsgContent 获取通知内容
func GenMsgContent(status, taskName, taskID string) (msgContent string) {
	h := "test.udns.woa.com"
	if conf.C().App.Mode == "prod" {
		h = "udns.woa.com"
	}
	switch status {
	case "doing":
		return fmt.Sprintf(
			"**udns流程待办提醒**\n\n>您有一个流程待办：%s\n"+
				">[点击查看详情](http://%s/#/flow/detail?task_id=%s)\n", taskName, h, taskID)
	case "done":
		return fmt.Sprintf(
			"**udns流程结束提醒**\n\n>您有一个流程已经结单：%s\n"+
				">[点击查看详情](http://%s/#/flow/detail?task_id=%s)\n", taskName, h, taskID)
	case "stop":
		return fmt.Sprintf(
			"**udns流程中止提醒**\n\n>您有一个流程已被中止：%s\n"+
				">[点击查看详情](http://%s/#/flow/detail?task_id=%s)\n", taskName, h, taskID)
	default:
		return ""
	}
}

// DoingTaskDailyNotify 定时任务通知
func (s *service) DoingTaskDailyNotify() error {
	a := make(map[string][]flow.Task)

	// 先获取还在doing的流程
	p := flow.QueryTaskParams{
		Page:     1,
		PageSize: 100,
		Status:   []string{"doing"},
	}
	res, err := s.QueryTask(flow.SystemUserName, &p)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"error": err,
		}).Error("query notify task error")
		return err
	}
	if res.Count > 100 {
		// todo 需要循环获取剩余的task_list
	}

	for _, i := range res.Results {
		t, err := s.DescribeTask(flow.SystemUserName, strconv.Itoa(int(i.ID)), true)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"id":    strconv.Itoa(int(i.ID)),
				"error": err,
			}).Error("describe task error")
			return err
		}
		for _, j := range t.Steps {
			if j.Status == flow.DoingStatus {
				if k, ok := a[j.Owner]; ok {
					// 这里需要注意，不能写成k = append(k, i)，这样无法赋值到map
					a[j.Owner] = append(k, i)
				} else {
					a[j.Owner] = []flow.Task{i}
				}
			}
		}
	}

	for user, f := range a {
		msg, err := GenMailContent(f)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"error": err,
			}).Error("GenMailContent error")
			return err
		}
		err = tool.SendMail(msg, user, "UDNS待办流程提醒")
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"msg":   msg,
				"error": err,
			}).Error("tool SendMail error")
			return err
		}
	}
	return nil
}

// GenMailContent 邮件模版
func GenMailContent(m []flow.Task) (msg string, err error) {
	buf := new(bytes.Buffer)

	// 模板渲染
	tmpl, err := template.ParseFiles("./template/mail.html")
	if err != nil {
		return
	}
	err = tmpl.Execute(buf, m)
	if err != nil {
		return "", err
	}
	msg = strings.Replace(buf.String(), "\n\n", "", -1)
	return
}
