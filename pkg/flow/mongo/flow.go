package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NopsTask Nops请求响应
type NopsTask struct {
	Data    *flow.Task `json:"data"`
	Msg     string     `json:"msg"`
	RetCode string     `json:"ret_code"`
}

// InsertTask 创建任务
func (s *service) InsertTask(id string) error {
	nopsHost := conf.C().Nops.WebapiHost
	url := fmt.Sprintf("%s/api/flow/qflow_task/%s", nopsHost, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"url":   url,
			"id":    id,
			"error": err,
		}).Error("http new request err:")
		return err
	}
	//req.Header.Set("Cookie", flow.TmpCookie)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"request": req,
			"error":   err,
		}).Error("do request err:")
		return err
	}

	defer resp.Body.Close()
	bodyC, _ := ioutil.ReadAll(resp.Body)

	var t NopsTask
	err = json.Unmarshal(bodyC, &t)
	if err != nil {
		s.l.Error("body unmarshal error:", err)
		return err
	}

	tx, err := s.col.InsertOne(context.TODO(), t.Data)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"data":  t.Data,
			"error": err,
		}).Error("insert task col error")
		return err
	}
	s.l.WithFields(logrus.Fields{
		"id":   tx.InsertedID,
		"data": t.Data,
	}).Info("insert task success")

	return nil
}

// QueryTask 查询任务
func (s *service) QueryTask(creator string, status string, stepOwner string, stepStatus []string, ps,
	pn int64) (count int64, rs []*flow.Task, err error) {
	//find option
	findOptions := options.Find()
	findOptions.SetLimit(ps)
	findOptions.SetSkip(ps * (pn - 1))
	findOptions.SetSort(bson.M{"start_time": -1})
	// filter 输入名称模糊匹配
	filter := bson.M{}
	if creator != "" {
		filter["creator"] = creator
		//filter["creator"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + name + ".*"}}
	}
	if status != "" {
		filter["status"] = status
	}
	if stepOwner != "" {
		filter["steps"] = bson.M{"$elemMatch": bson.M{"status": bson.M{"$in": stepStatus},
			"owner": bson.M{"$regex": primitive.Regex{Pattern: stepOwner + "(;|$)"}}}}
	}

	count, err = s.col.CountDocuments(context.TODO(), filter)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"filter": filter,
			"error":  err,
		}).Error("query task count error")
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"creator": creator,
			"filter":  filter,
			"error":   err,
		}).Error("query flows error")
		return
	}

	for cur.Next(context.TODO()) {
		var t flow.Task
		err = cur.Decode(&t)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"creator": creator,
				"error":   err,
			}).Error("decode r err when query flows")
			return 0, nil, err
		}
		rs = append(rs, &t)
	}
	return
}
