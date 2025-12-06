package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/zone"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertZone 添加zone
func (s *service) InsertZone(z *zone.Zone) error {
	ok, err := s.IsZoneExisted(z.Name)
	if err != nil {
		s.l.Error("check zone existed error:", err)
		return err
	}
	if ok {
		s.l.Error("zone already existed:", z.Name)
		return errors.New("zone already existed")
	}
	// 初始值
	z.ID = primitive.NewObjectID()
	z.CreatedTime = time.Now().Unix()
	z.UpdatedTime = 0
	z.Deleted = false

	tx, err := s.col.InsertOne(context.TODO(), z)
	if err != nil {
		s.l.Error("insert zone col error:", err)
		return err
	}
	s.l.WithFields(logrus.Fields{
		"_id":      tx.InsertedID,
		"name":     z.Name,
		"operator": z.Operator,
		"enabled":  z.Enabled,
	}).Info("insert zone success")
	return nil
}

// DeleteZone 删除zone
func (s *service) DeleteZone(id string) error {
	now := time.Now().Unix()
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.Error("soft delete zone col error:", err)
		return err
	}
	d := zone.Zone{}
	if err = s.col.FindOne(context.TODO(), filter).Decode(&d); err != nil {
		return err
	}
	if tx.MatchedCount == 0 {
		s.l.Error("delete fail: can not find zone by id")
		return errors.New("delete fail: can not find zone by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":    id,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete zone success")
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(d.Name) {
			if err = cache.C().Delete(d.Name); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": d.Name,
				}).Error("delete cache error:", err)
			}
		}
	}
	return nil
}

// UpdateZone 更新zone
func (s *service) UpdateZone(z *zone.Zone) error {
	now := time.Now().Unix()
	filter := bson.M{"_id": z.ID, "deleted": false}
	update := bson.M{"$set": bson.M{"name": z.Name, "operator": z.Operator,
		"enabled": z.Enabled, "description": z.Description, "zone_id": z.ZoneID, "updated_time": now}}

	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.Error("Update zone col error:", err)
		return err
	}
	if tx.MatchedCount == 0 {
		s.l.Error("update fail: can not find zone by id")
		return errors.New("update fail: can not find zone by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id": z.ID,

		"operator":    z.Operator,
		"enabled":     z.Enabled,
		"description": z.Description,
		"match":       tx.MatchedCount,
		"delete":      tx.ModifiedCount,
	}).Info("update zone success")
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(z.Name); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": z.Name,
			}).Error("delete cache error:", err)
		}
	}
	return nil
}

// QueryZones 查询zone
func (s *service) QueryZones(name string, ps, pn int64) (count int64, zs []*zone.Zone, err error) {
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
		}).Error("find zone count error:", err)
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("query zones err:", err)
		return
	}

	for cur.Next(context.TODO()) {
		var z zone.Zone
		err = cur.Decode(&z)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"name":  name,
				"error": err,
			}).Error("decode z err when query zones:", err)
			return 0, nil, err
		}
		zs = append(zs, &z)
	}
	return
}

// DescribeZone 查询zone
func (s *service) DescribeZone(name string) (d *zone.Zone, err error) {
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(name) {
			if err = cache.C().Get(name, &d); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": name,
				}).Error("get cache error:", err)
			} else {
				return
			}
		}
	}
	filter := bson.M{"name": name, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(&d); err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("find zone col error:", err)
		return nil, err
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(name, d, domain.HalfHourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": name,
			}).Error("set cache error:", err)
		}
	}
	return d, nil
}

// IsZoneExisted 判断zone是否存在
func (s *service) IsZoneExisted(name string) (ok bool, err error) {
	e := &zone.Zone{}
	filter := bson.M{"name": name, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
