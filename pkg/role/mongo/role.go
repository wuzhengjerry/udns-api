package mongo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertRole 添加角色
func (s *service) InsertRole(r *role.Role) error {
	ok, err := s.IsRoleExisted(r.Name)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  r.Name,
			"users": r.Users,
			"error": err,
		}).Error("check role existed error")
		return err
	}
	if ok {
		return errors.New("role already existed")
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
			"users": r.Users,
			"error": err,
		}).Error("insert role error")
		return err
	}

	s.l.WithFields(logrus.Fields{
		"_id":   tx.InsertedID,
		"name":  r.Name,
		"users": r.Users,
	}).Info("insert role success")
	return nil
}

// DeleteRole 删除角色
func (s *service) DeleteRole(name string) error {
	now := time.Now().Unix()
	filter := bson.M{"name": name, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("soft delete role error")
		return err
	}
	if tx.MatchedCount == 0 {
		return errors.New("delete fail: can not find role by id")
	}
	s.l.WithFields(logrus.Fields{
		"name":   name,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete role success")
	// 删除缓存，但这里一般不做删除角色操作
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(name); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": name,
			}).Error("delete cache error:", err)
		}
	}
	return nil
}

// UpdateRole 更新角色
func (s *service) UpdateRole(r *role.Role) error {
	now := time.Now().Unix()
	filter := bson.M{"_id": r.ID, "deleted": false}
	update := bson.M{"$set": bson.M{"users": r.Users, "description": r.Description, "updated_time": now}}

	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"users": r.Users,
			"error": err,
		}).Error("update role error")
		return err
	}
	if tx.MatchedCount == 0 {
		return errors.New("update fail: can not find Role by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":         r.ID,
		"name":        r.Name,
		"users":       r.Users,
		"description": r.Description,
		"match":       tx.MatchedCount,
		"delete":      tx.ModifiedCount,
	}).Info("update Role success")
	// 删除缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(string(r.Name)); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": string(r.Name),
			}).Error("delete cache error:", err)
		}
	}
	return nil
}

// QueryRoles 角色列表
func (s *service) QueryRoles(name string, ps, pn int64) (count int64, rs []*role.Role, err error) {
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
		}).Error("query roles count error")
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("query roles err:", err)
		return
	}

	for cur.Next(context.TODO()) {
		var r role.Role
		err = cur.Decode(&r)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"name":  name,
				"error": err,
			}).Error("decode r err when query roles:", err)
			return 0, nil, err
		}
		rs = append(rs, &r)
	}
	return
}

// DescribeRole 角色信息
func (s *service) DescribeRole(name role.RoleName) (r *role.Role, err error) {
	// 首先从缓存中获取
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(string(name)) {
			if err = cache.C().Get(string(name), &r); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": string(name),
				}).Error("get cache error:", err)
			} else {
				return
			}
		}
	}

	filter := bson.M{"name": name, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(&r); err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("describe role err:", err)
		return nil, err
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(string(name), r, role.HalfHourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": string(name),
			}).Error("set cache error:", err)
		}
	}
	return r, nil
}

// IsRoleExisted 角色是否存在
func (s *service) IsRoleExisted(name role.RoleName) (ok bool, err error) {
	e := &role.Role{}
	filter := bson.M{"name": name, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// HasRolePermission 判断用户是否属于某个角色
func (s *service) HasRolePermission(roleName role.RoleName, user string) bool {
	r, err := s.DescribeRole(roleName)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":     user,
			"roleName": roleName,
			"error":    err,
		}).Error("auth user has role perm error")
		return false
	}
	userList := strings.Split(r.Users, ";")
	for _, j := range userList {
		if user == j {
			return true
		}
	}
	return false
}
