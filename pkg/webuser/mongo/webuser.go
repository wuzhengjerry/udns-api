package mongo

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/webuser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IsUserExisted 用户是否存在
func (s *service) IsUserExisted(name string) (ok bool, err error) {
	filter := bson.M{"name": name, "deleted": false}
	e := &webuser.WebUser{}
	if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Sha1DigestBase64 加密参数
func (s *service) Sha1DigestBase64(body []byte) string {
	sha := md5.New()
	sha.Write(body)
	return fmt.Sprintf("MD5=%s", base64.StdEncoding.EncodeToString(sha.Sum(nil)))
}

// HmacSHA1 生成签名
func (s *service) HmacSHA1(key string, data string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	result := hex.EncodeToString(mac.Sum(nil))
	return strings.ToUpper(result)
}

// InsertUser 添加用户
func (s *service) InsertUser(d *webuser.WebUser) error {
	ok, err := s.IsUserExisted(d.Name)
	if err != nil {
		s.l.Error("check user existed err:", err)
		return err
	}
	if ok {
		return errors.New("user already existed")
	}
	// 初始值
	body, err := json.Marshal(&d)
	if err != nil {
		s.l.Error("json marshal web user body err:", err)
		return err
	}
	bodyDigest := s.Sha1DigestBase64(body)
	gmTime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT") //获取时间字符串
	//拼装待签名的数据
	strToSign := fmt.Sprintf("date: %s\ndigest: %s", gmTime, bodyDigest)
	s.l.Info("strToSign", strToSign)
	//生成签名
	sign := s.HmacSHA1(d.Name, strToSign)
	s.l.Info("sign", sign)
	var domains []string
	for _, item := range d.Domain {
		tmp := item
		if item[len(item)-1:] != "." {
			tmp = tmp + "."
		}
		domains = append(domains, tmp)
	}
	d.ID = primitive.NewObjectID()
	d.CreatedTime = time.Now().Unix()
	d.UpdatedTime = 0
	d.Deleted = false
	d.Token = sign
	d.Domain = domains
	tx, err := s.col.InsertOne(context.TODO(), d)
	if err != nil {
		s.l.Error("insert web user col err:", err)
		return err
	}
	info := fmt.Sprintf("create user success, token is: %s", sign)
	s.l.WithFields(logrus.Fields{
		"_id":         tx.InsertedID,
		"name":        d.Name,
		"operator":    d.Operator,
		"description": d.Description,
		"token":       sign,
	}).Info(info)
	return nil
}

// DeleteUser 删除用户
func (s *service) DeleteUser(name string) error {
	now := time.Now().Unix()
	filter := bson.M{"name": name, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name": name,
		}).Error("soft delete user error:", err)
		return err
	}
	if tx.MatchedCount == 0 {
		s.l.WithFields(logrus.Fields{
			"name": name,
		}).Error("can not find user by name")
		return errors.New("delete fail: can not find user by name")
	}
	s.l.WithFields(logrus.Fields{
		"name":   name,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete user success")
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(name) {
			if err = cache.C().Delete(name); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": name,
				}).Error("delete cache error:", err)
			}
		}
	}
	return nil
}

// UpdateUser 更新用户
func (s *service) UpdateUser(z *webuser.WebUser) error {
	now := time.Now().Unix()
	filter := bson.M{"name": z.Name, "deleted": false}
	update := bson.M{"$set": bson.M{"name": z.Name, "operator": z.Operator,
		"domain": z.Domain, "description": z.Description, "updated_time": now}}

	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name": z.Name,
		}).Error("update user error:", err)
		return err
	}
	if tx.MatchedCount == 0 {
		s.l.WithFields(logrus.Fields{
			"name": z.Name,
		}).Error("can not find user by id")
		return errors.New("update fail: can not find user by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":         z.ID,
		"name":        z.Name,
		"operator":    z.Operator,
		"description": z.Description,
		"match":       tx.MatchedCount,
		"delete":      tx.ModifiedCount,
	}).Info("update user success")
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(z.Name); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": z.Name,
			}).Error("delete cache error:", err)
		}
	}
	return nil
}

// QueryUsers 查询用户
func (s *service) QueryUsers(name string, ps, pn int64) (count int64, zs []*webuser.WebUser, err error) {
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
		s.l.Error("query user count error:", err)
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("query users err:", err)
		return
	}

	for cur.Next(context.TODO()) {
		var z webuser.WebUser
		err = cur.Decode(&z)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"name":  name,
				"error": err,
			}).Error("decode z err when query users:", err)
			return 0, nil, err
		}
		zs = append(zs, &z)
	}
	return
}

// DescribeUser 用户信息
func (s *service) DescribeUser(name string) (d *webuser.WebUser, err error) {
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(name) {
			if err = cache.C().Get(name, &d); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": name,
				}).Error("get cache error:", err)
			} else {
				return d, nil
			}
		}
	}
	filter := bson.M{"name": name, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(&d); err != nil {
		s.l.WithFields(logrus.Fields{
			"name": name,
		}).Error("describe user error:", err)
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
