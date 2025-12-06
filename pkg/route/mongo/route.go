package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/pkg/route"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertRoute 添加路由
func (s *service) InsertRoute(r *route.Route) error {
	ok, err := s.IsRouteExisted(r.RouteID)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  r.Name,
			"error": err,
		}).Error("check route existed error")
		return err
	}
	if ok {
		return errors.New(fmt.Sprintf("route_id %s already existed", r.RouteID))
	}
	// 初始值
	r.ID = primitive.NewObjectID()
	r.CreatedTime = time.Now().Unix()
	r.UpdatedTime = 0
	r.Deleted = false

	tx, err := s.col.InsertOne(context.TODO(), r)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  r.Name,
			"error": err,
		}).Error("insert route col error")
		return err
	}
	s.l.WithFields(logrus.Fields{
		"_id":      tx.InsertedID,
		"name":     r.Name,
		"route_id": r.RouteID,
		"operator": r.Operator,
		"enabled":  r.Enabled,
	}).Info("insert route success")
	return nil
}

// DeleteRoute 删除路由
func (s *service) DeleteRoute(id string) error {
	now := time.Now().Unix()
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"_id":   objID,
			"error": err,
		}).Error("soft delete route error")
		return err
	}
	if tx.MatchedCount == 0 {
		return errors.New("delete fail: can not find Route by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":    id,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete route success")
	return nil
}

// UpdateRoute 更新路由
func (s *service) UpdateRoute(r *route.Route) error {
	now := time.Now().Unix()
	filter := bson.M{"_id": r.ID, "deleted": false}
	update := bson.M{"$set": bson.M{"operator": r.Operator, "route_id": r.RouteID, "name": r.Name,
		"enabled": r.Enabled, "description": r.Description, "updated_time": now}}

	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"_id":   r.ID,
			"name":  r.Name,
			"error": err,
		}).Error("update route error")
		return err
	}
	if tx.MatchedCount == 0 {
		return errors.New("update fail: can not find route by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":         r.ID,
		"name":        r.Name,
		"route_id":    r.RouteID,
		"operator":    r.Operator,
		"enabled":     r.Enabled,
		"description": r.Description,
		"match":       tx.MatchedCount,
		"delete":      tx.ModifiedCount,
	}).Info("update route success")
	return nil
}

// QueryRoutes 路由列表
func (s *service) QueryRoutes(name string, ps, pn int64) (count int64, rs []*route.Route, err error) {
	findOptions := options.Find()
	findOptions.SetLimit(ps)
	findOptions.SetSkip(ps * (pn - 1))
	findOptions.SetSort(bson.M{"created_time": -1})
	// filter 输入名称模糊匹配
	filter := bson.M{"deleted": false}
	if name != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + name + ".*"}}
	}

	count, err = s.col.CountDocuments(context.TODO(), filter)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("query routes count error")
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("query routes err:", err)
		return
	}

	for cur.Next(context.TODO()) {
		var r route.Route
		err = cur.Decode(&r)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"name":  name,
				"error": err,
			}).Error("decode r err when query Routes:", err)
			return 0, nil, err
		}
		rs = append(rs, &r)
	}
	return
}

// DescribeRoute 路由信息
func (s *service) DescribeRoute(id string) (*route.Route, error) {
	a := route.Route{}
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "deleted": false}
	if err := s.col.FindOne(context.TODO(), filter).Decode(&a); err != nil {
		s.l.WithFields(logrus.Fields{
			"_id":   objID,
			"error": err,
		}).Error("describe route error")
		return nil, err
	}
	return &a, nil
}

// IsRouteExisted 路由是否存在
func (s *service) IsRouteExisted(id string) (ok bool, err error) {
	e := &route.Route{}
	filter := bson.M{"route_id": id, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
