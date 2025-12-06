package mongo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/tool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InsertDomain 新增域名
func (s *service) InsertDomain(user string, d *domain.Domain) error {
	ok, err := s.IsDomainExisted(d.Name)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"error": err,
		}).Error("check domain exist error")
		return err
	}
	if ok {
		return errors.New("domain already existed")
	}
	// 初始值
	d.ID = primitive.NewObjectID()
	d.CreatedTime = time.Now().Unix()
	d.UpdatedTime = 0
	d.Deleted = false
	tx, err := s.col.InsertOne(context.TODO(), d)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"data":  d,
			"user":  user,
			"error": err,
		}).Error("insert domain col error")
		return err
	}
	params, fErr := json.Marshal(d)
	if fErr != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"error": fErr,
		}).Error("create domain param formatter error")
	}
	s.SyncUdnsApiLog(user, string(params), d.Name, "udns_web", "申请", "是", "success")

	s.l.WithFields(logrus.Fields{
		"_id":         tx.InsertedID,
		"name":        d.Name,
		"owner":       user,
		"operator":    d.Operator,
		"business_id": d.BusinessID,
		"group_id":    d.GroupID,
	}).Info("insert domain success")
	return nil
}

// DeleteDomain 删除域名
func (s *service) DeleteDomain(user, name string) (err error) {
	now := time.Now().Unix()
	filter := bson.M{"name": name, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"user":  user,
			"error": err,
		}).Error("soft delete domain col error")
		return
	}

	if tx.MatchedCount == 0 {
		return errors.New("delete fail: can not find domain by name")
	}

	s.SyncUdnsApiLog(user, name, name, "udns_web", "删除", "是", "success")

	s.l.WithFields(logrus.Fields{
		"user":   user,
		"name":   name,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete domain success")

	if conf.C().Cache.IsCache {
		if cache.C().IsExist(name) {
			if err = cache.C().Delete(name); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": name,
					"error":     err,
				}).Error("delete cache error")
			}
		}
	}
	return
}

// UpdateDomain 更新域名
func (s *service) UpdateDomain(user string, d *domain.Domain) error {
	groupId := d.GroupID
	groupName := d.GroupName
	deptId := d.DeptID
	deptName := d.DeptName
	userInfo, err := s.GetStaffInfo(d.Owner)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"domain": d,
			"error":  err,
		}).Error("get owner staff info error")
	} else {
		if userInfo["GroupId"] != nil && userInfo["DeptId"] != nil {
			groupId = int(userInfo["GroupId"].(float64))
			groupName = fmt.Sprintf("%s-%s", userInfo["DepartmentName"].(string),
				userInfo["GroupName"].(string))
			deptId = int(userInfo["DeptId"].(float64))
			deptName = userInfo["DepartmentName"].(string)
		}
	}
	now := time.Now().Unix()
	filter := bson.M{"name": d.Name, "deleted": false}
	update := bson.M{"$set": bson.M{"name": d.Name, "owner": d.Owner, "operator": d.Operator, "business_id": d.BusinessID,
		"group_id": groupId, "business_name": d.BusinessName, "group_name": groupName, "dept_id": deptId,
		"dept_name": deptName, "description": d.Description, "updated_time": now}}
	tx, err := s.col.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"update": update,
			"error":  err,
		}).Error("update domain error")
		return err
	}
	if tx.MatchedCount == 0 {
		return errors.New("update fail: can not find domain by name")
	}
	s.l.WithFields(logrus.Fields{
		"_id":         d.ID,
		"name":        d.Name,
		"owner":       d.Owner,
		"operator":    d.Operator,
		"business_id": d.BusinessID,
		"group_id":    d.GroupID,
		"match":       tx.MatchedCount,
		"delete":      tx.ModifiedCount,
	}).Info("update Domain success")

	params, _ := json.Marshal(d)
	s.SyncUdnsApiLog(user, string(params), d.Name, "udns_web", "变更", "是", "success")

	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(d.Name); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": d.Name,
				"error":     err,
			}).Error("delete cache error")
		}
	}
	return nil
}

// QueryDomains 域名列表
func (s *service) QueryDomains(
	manage, name, owner, operator, description, businessId string,
	ps, pn int64) (count int64, rs []*domain.Domain, err error) {
	findOptions := options.Find()
	findOptions.SetLimit(ps)
	findOptions.SetSkip(ps * (pn - 1))
	findOptions.SetSort(bson.M{"created_time": -1})
	filter := bson.M{"deleted": false}
	if name != "" { //忽略大小写'
		//屏蔽特殊符号"*"和"."对查询结果的影响
		tmp := strings.ReplaceAll(name, ".", "\\.")
		tmp = strings.ReplaceAll(tmp, "*", "\\*")
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + tmp + ".*", Options: "i"}} //"i是忽略大小写"
	}
	if operator != "" {
		filter["operator"] = bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + operator + "(;|$)"}}
	}
	if owner != "" {
		filter["owner"] = owner
	}
	if description != "" {
		filter["description"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + description + ".*"}}
	}
	if businessId != "" {
		filter["business_id"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + businessId + ".*"}}
	}
	if manage != "" { //查询用户管理的域名
		filter["$or"] = []bson.M{
			bson.M{"owner": manage},
			bson.M{"operator": bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + manage + "(;|$)"}}}}
	}
	count, err = s.col.CountDocuments(context.TODO(), filter)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"owner":    owner,
			"name":     name,
			"operator": operator,
			"error":    err,
		}).Error("query domain count error")
		return
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"owner":    owner,
			"name":     name,
			"operator": operator,
			"error":    err,
		}).Error("query domains error")
		return
	}

	for cur.Next(context.TODO()) {
		var d domain.Domain
		err = cur.Decode(&d)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"owner":    owner,
				"name":     name,
				"operator": operator,
				"error":    err,
			}).Error("decode d err when query domains")
			return 0, nil, err
		}
		rs = append(rs, &d)
	}
	return
}

// DescribeDomain 获取域名信息
func (s *service) DescribeDomain(domainName string) (domainInfo *domain.Domain, err error) {
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(domainName) {
			if err = cache.C().Get(domainName, &domainInfo); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": domainName,
					"error":     err,
				}).Error("get cache error")
			} else {
				return
			}
		}
	}
	filter := bson.M{"name": domainName, "deleted": false}
	if err = s.col.FindOne(context.TODO(), filter).Decode(&domainInfo); err != nil {
		s.l.WithFields(logrus.Fields{
			"domainName": domainName,
			"error":      err,
		}).Error("mongodb FindOne domain error")
		return domainInfo, err
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(domainName, domainInfo, domain.HalfHourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": domainName,
				"error":     err,
			}).Error("set cache error")
		}
	}
	return domainInfo, nil
}

// HasDomainPermission 判断用户域名操作权限
func (s *service) HasDomainPermission(name, user string) bool {
	r, err := s.DescribeDomain(name)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"name":  name,
			"error": err,
		}).Info("user has domain perm error")
		return false
	}
	operatorList := strings.Split(r.Operator, ";")
	owner := r.Owner
	if owner == user {
		return true
	} else {
		for _, j := range operatorList {
			if user == j {
				return true
			}
		}
	}
	return false
}

// InsertDomainLog 域名操作日志
func (s *service) InsertDomainLog(user string, d *domain.DomainLog) error {
	d.ID = primitive.NewObjectID()
	d.CreatedTime = time.Now().Unix()
	d.UpdatedTime = 0
	d.Deleted = false
	filter := bson.M{"deleted": false, "domain_name": d.DomainName}
	count, err := s.logs.CountDocuments(context.TODO(), filter)
	d.Version = count + 1
	tx, err := s.logs.InsertOne(context.TODO(), d)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"data":  d,
			"error": err,
		}).Error("insert domain log error")
		return err
	}
	s.l.WithFields(logrus.Fields{
		"_id":         tx.InsertedID,
		"domain_name": d.DomainName,
		"operator":    user,
	}).Info("insert domain log success")
	return nil
}

// InsertSyncApiLog 同步IT,udns_web, udns_ori,udns_openapi操作日志
func (s *service) InsertSyncApiLog(user string, d *domain.SyncApiLog) error {
	d.ID = primitive.NewObjectID()
	d.CreatedTime = time.Now().Unix()
	d.UpdatedTime = 0
	d.Deleted = false
	tx, err := s.itlogs.InsertOne(context.TODO(), d)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"data":  d,
			"error": err,
		}).Error("insert api log error")
		return err
	}
	s.l.WithFields(logrus.Fields{
		"_id":         tx.InsertedID,
		"domain_name": d.DomainName,
		"api_type":    d.ApiType,
		"operator":    user,
	}).Info("insert domain log success")
	return nil
}

// DeleteDomainLog 删除域名操作日志
func (s *service) DeleteDomainLog(id string) (err error) {
	now := time.Now().Unix()
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true, "deleted_time": now}}
	// 这里做软删除
	tx, err := s.logs.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"_id":   id,
			"error": err,
		}).Error("soft delete domain log error")
		return
	}
	if tx.MatchedCount == 0 {
		return errors.New("delete fail: can not find domain by id")
	}
	s.l.WithFields(logrus.Fields{
		"_id":    id,
		"match":  tx.MatchedCount,
		"delete": tx.ModifiedCount,
	}).Info("soft delete domain success")
	return
}

// QueryDomainLogs 查询域名操作日志列表
func (s *service) QueryDomainLogs(name string, ps, pn int64) (count int64, rs []*domain.DomainLog, err error) {
	findOptions := options.Find()
	findOptions.SetLimit(ps)
	findOptions.SetSkip(ps * (pn - 1))
	findOptions.SetSort(bson.M{"created_time": -1})
	// filter 输入名称模糊匹配
	filter := bson.M{"deleted": false}
	if name != "" {
		filter["domain_name"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + name + ".*"}}
	}
	count, err = s.logs.CountDocuments(context.TODO(), filter)
	if err != nil {
		return
	}
	cur, err := s.logs.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"domain_name": name,
			"error":       err,
		}).Error("query domains err:", err)
		return
	}

	for cur.Next(context.TODO()) {
		var d domain.DomainLog
		err = cur.Decode(&d)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"domain_name": name,
				"error":       err,
			}).Error("decode d err when query domains:", err)
			return 0, nil, err
		}
		rs = append(rs, &d)
	}
	return
}

// QuerySyncApiLogs 查看域名同步日志列表
func (s *service) QuerySyncApiLogs(name,
	user, method, apiType, success string, ps, pn int64) (count int64, rs []*domain.SyncApiLog, err error) {
	findOptions := options.Find()
	findOptions.SetLimit(ps)
	findOptions.SetSkip(ps * (pn - 1))
	findOptions.SetSort(bson.M{"created_time": -1})
	// filter 输入名称模糊匹配
	filter := bson.M{"deleted": false}
	if name != "" {
		tmp := strings.ReplaceAll(name, ".", "\\.")
		tmp = strings.ReplaceAll(tmp, "*", "\\*")
		filter["domain_name"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + tmp + ".*", Options: "i"}} //"i是忽略大小写"
		//filter["domain_name"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + name + ".*"}, "option": "si"}
	}
	if user != "" {
		filter["operator"] = user
	}
	if method != "" {
		filter["method"] = method
	}
	if apiType != "" {
		filter["api_type"] = apiType
	}
	if success != "" {
		filter["success"] = success
	}
	count, err = s.itlogs.CountDocuments(context.TODO(), filter)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"filter":      filter,
			"domain_name": name,
			"error":       err,
		}).Error("query it log count err:", err)
		return
	}
	cur, err := s.itlogs.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"domain_name": name,
			"error":       err,
		}).Error("query it log err:", err)
		return
	}
	for cur.Next(context.TODO()) {
		var d domain.SyncApiLog
		err = cur.Decode(&d)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"domain_name": name,
				"error":       err,
			}).Error("decode d err when query sync it log:", err)
			return 0, nil, err
		}
		rs = append(rs, &d)
	}
	return
}

// DescribeDomainLog 获取域名日志信息
func (s *service) DescribeDomainLog(id string) (*domain.DomainLog, error) {
	res := domain.DomainLog{}
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "deleted": false}
	if err := s.logs.FindOne(context.TODO(), filter).Decode(&res); err != nil {
		s.l.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("describe domain log error")
		return nil, err
	}
	return &res, nil
}

// IsDomainsExisted 判断域名是否存在
func (s *service) IsDomainsExisted(names string) (res map[string]interface{}, err error) { //鉴定域名是否存在，是否以二级zone结尾
	allNames := strings.Split(names, ";")
	result := make(map[string]bool) //记录存在的域名
	e := &domain.Domain{}
	msg := ""
	tmpData := make(map[string]interface{})
	res = tmpData
	for _, j := range allNames {
		tmp := j
		if tmp[len(j)-1:] != "." {
			tmp = strings.TrimSpace(tmp) + "."
		} //验证是否以.结尾,没有则要加"点"，默认DB存储域名加点
		filter := bson.M{"name": tmp, "deleted": false}
		if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				res["Message"] = fmt.Sprintf("%s", err)
				res["Success"] = false
			}
		} else {
			result[j] = true
		}
	}
	if len(result) != 0 {
		for key, _ := range result {
			msg = msg + key + ";"
		}
		res["Message"] = msg + "已存在"
		res["Success"] = true
		return res, nil
	}
	res["Message"] = "Non Existed"
	res["Success"] = false
	return res, nil
}

// IsOADomainExisted 判断oa域名是否存在
func (s *service) IsOADomainExisted(user, name string) (res domain.Message, err error) {
	name = s.SplitStr(name)
	resp, err := s.RequestRITApi(name)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": err,
		}).Error("queryITData error")
		res.Message = fmt.Sprintf("queryITData error: %s", err)
		res.Success = false
	}
	if resp.DomainName != "" {
		//获取到OA域名
		curUser := strings.Split(resp.Admin, ";")
		flag := false
		if len(curUser) > 0 {
			for _, u := range curUser {
				if user == u {
					flag = true
					break
				}
			}
			if !flag {
				res.Message = fmt.Sprintf("IT exists domain:%s, admin: %s, %s is forbidden to apply woa domain",
					name, resp.Admin, user)
				res.Success = false
			} else {
				res.Success = true
				res.Message = fmt.Sprintf("get %s domain successfully, %s has permission", name, user)
			}
		} else {
			res.Success = true
			res.Message = fmt.Sprintf("get %s domain successfully, admin is idle", name)
		}

	} else {
		//获取oa域名空
		res.Success = true
		res.Message = fmt.Sprintf("get %s domain is idle", name)
	}
	return
}

// QueryITDomainsExisted 判断IT侧域名是否存在
func (s *service) QueryITDomainsExisted(user, names string) (res domain.Message, err error) {
	//鉴定IT中oa.com域名是否存在，申请用户是否有权限申请woa.com
	allNames := strings.Split(names, ";")
	result := make(map[string]domain.Message) //记录存在的域名
	msg := ""
	var swg sync.WaitGroup
	for _, j := range allNames {
		swg.Add(1)
		go func(wg *sync.WaitGroup, mark map[string]domain.Message, user, j string) {
			curRes, curErr := s.IsOADomainExisted(user, j)
			resp := domain.Message{Success: true, Message: "Success"}
			if curErr != nil {
				resp.Message = fmt.Sprintf("query %s domain error: %s", j, curErr)
				resp.Success = false
				curRes = resp
				s.l.WithFields(logrus.Fields{
					"user":    user,
					"names":   names,
					"Message": resp.Message,
					"error":   curErr,
				}).Error("query OA domain error")
			}
			mark[j] = curRes
			defer wg.Done() //等价于 wg.Add(-1)
		}(&swg, result, user, j)
	}
	swg.Wait()
	finalRes := 0
	for key, value := range result {
		fmt.Println(key, value)
		if !value.Success {
			finalRes += 1
			if msg != "" {
				msg = fmt.Sprintf("%s; %s", msg, value.Message)
			} else {
				msg = fmt.Sprintf("%s", value.Message)
			}
		}
	}
	res.Message = msg
	if finalRes > 0 {
		res.Success = false
	} else {
		res.Success = true
	}
	return res, nil
}

// IsDomainExisted 域名是否存在
func (s *service) IsDomainExisted(name string) (ok bool, err error) {
	filter := bson.M{"name": name, "deleted": false}
	e := &domain.Domain{}
	if err = s.col.FindOne(context.TODO(), filter).Decode(e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetAllStaffFullName 获取公司全部用户
func (s *service) GetAllStaffFullName() (res string, err error) {
	//先从缓存中获取
	keyName := "get_all_staff_fullname"
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(keyName) {
			if err = cache.C().Get(keyName, &res); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": keyName,
					"error":     err,
				}).Error("get cache error")
			} else {
				return
			}
		}
	}
	sniperHost := conf.C().TOF.SniperHost
	url := sniperHost + "/script_service/get_all_staff_fullname"
	client := &http.Client{Timeout: 60 * time.Second} // 1min超时
	resp, err := client.Get(url)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"url":   url,
			"error": err,
		}).Error("get all staff full name error")
		return "", err
	}
	defer resp.Body.Close()
	var buffer [2048]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err1 := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err1 != nil && err1 == io.EOF {
			break
		} else if err1 != nil {
			s.l.WithFields(logrus.Fields{
				"result": result.String(),
				"error":  err1,
			}).Error("response body read error")
			return "", err1
		}
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(keyName, result.String(), domain.HourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": keyName,
			}).Error("set cache error:", err)
		}
	}
	return result.String(), nil
}

// GetStaffInfo 获取员工信息
func (s *service) GetStaffInfo(user string) (res map[string]interface{}, err error) { //根据用户名获取部门信息
	isSuper := false
	if pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		isSuper = true
	}
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(user) {
			if err = cache.C().Get(user, &res); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": user,
				}).Error("get cache error:", err)
			} else {
				res["is_super"] = isSuper
				return
			}
		}
	}
	gslbHost := conf.C().TOF.GSLBHost
	url := fmt.Sprintf("%s/test/staff/%s/json", gslbHost, user)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"url":   url,
			"error": err,
		}).Error("http NewRequest error")
		return nil, err
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"req":   req,
			"error": err,
		}).Error("do request error")
		return nil, err
	}
	defer resp.Body.Close()
	tmpRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"body":  resp.Body,
			"error": err,
		}).Error("body read error")
		return nil, err
	}
	err = json.Unmarshal(tmpRes, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   string(tmpRes),
			"error": err,
		}).Error("body json unmarshal error")
		return nil, err
	}
	res["is_super"] = isSuper
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(user, res, domain.HourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": user,
			}).Error("set cache error:", err)
		}
	}
	return
}

// GetBusinessTree 获取业务树
func (s *service) GetBusinessTree() (res map[string]interface{}, err error) {
	keyName := "get_business_tree"
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(keyName) {
			if err = cache.C().Get("get_business_tree", &res); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": keyName,
				}).Error("get cache error:", err)
			} else {
				return
			}
		}
	}
	nopsHost := conf.C().Nops.WebapiHost
	url := nopsHost + "/api/cmdb/business/business_tree?format=1&use_cache=1&level=3&with_set=0"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"url":   url,
			"error": err,
		}).Error("http NewRequest error")
		return nil, err
	}
	req.Header.Set("Cookie", domain.TmpCookie)
	req.Header.Set("BusinessID", "*")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"req":   req,
			"error": err,
		}).Error("do request error")
		return nil, err
	}
	defer resp.Body.Close()
	tmpRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"body":  resp.Body,
			"error": err,
		}).Error("body read error")
		return nil, err
	}
	err = json.Unmarshal(tmpRes, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   string(tmpRes),
			"error": err,
		}).Error("body json unmarshal error")
		return nil, err
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(keyName, res, domain.HourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": keyName,
			}).Error("set cache error:", err)
		}
	}
	return
}

// GetDomainConfig 获取域名配置
func (s *service) GetDomainConfig(params *domain.QueryDomainParam) (res interface{}, err error) {
	params.NSysId = conf.C().OUDNS.SystemID
	key := fmt.Sprintf("udns_ori_%s", params.SDomain)
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(key) {
			if err = cache.C().Get(key, &res); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": key,
				}).Error("get cache error:", err)
			} else {
				return
			}
		}
	}
	resp, err := RequestUdns(conf.C().OUDNS.QueryApi, http.MethodPost, params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("domain config unmarshal error")
		return nil, err
	}
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(key, res, domain.HalfHourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": key,
			}).Error("set cache error:", err)
		}
	}
	return res, nil
}

// BatchGetDomainConfig 批量获取域名配置
func (s *service) BatchGetDomainConfig(params *domain.BatchQueryDomainParam) (res []*domain.ModDomainParam, err error) {
	params.NSysId = conf.C().OUDNS.SystemID
	resp, err := RequestUdns(conf.C().OUDNS.BatchQueryApi, http.MethodPost, params)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(resp, &result)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("batch domain config unmarshal error")
		return
	}
	curConfig, err := json.Marshal(result["DomainInfo"])
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("DomainInfo marshal error")
		return
	}
	err = json.Unmarshal(curConfig, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"curConfig": string(curConfig),
			"error":     err,
		}).Error("DomainInfo unmarshal error")
		return nil, err
	}
	return
}

// DeleteDomainConfig 删除域名配置
func (s *service) DeleteDomainConfig(user string, params *domain.DelDomainParam) (res interface{}, err error) {
	if flag := s.CheckUserDomainPermission(user, params.SDomain); flag == false {
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"domain": params.SDomain,
		}).Error("domain permission deny")
		return res, errors.New("permission deny")
	}
	return s.DeleteDomainConfigAction(params)
}

// DeleteDomainConfigAction 删除域名配置操作
func (s *service) DeleteDomainConfigAction(params *domain.DelDomainParam) (res interface{}, err error) {
	params.NSysId = conf.C().OUDNS.SystemID
	params.SProposer = conf.C().OUDNS.Proposer
	key := fmt.Sprintf("udns_ori_%s", params.SDomain)
	resp, err := RequestUdns(conf.C().OUDNS.DelApi, http.MethodPost, params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("udns resp unmarshal error")
		return nil, err
	}
	if conf.C().Cache.IsCache {
		if cacheErr := cache.C().Delete(key); cacheErr != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": key,
			}).Error("delete cache error:", cacheErr)
		}
	}
	return
}

// AddDomainConfig 添加域名配置
func (s *service) AddDomainConfig(params *domain.ModDomainParam) (res interface{}, err error) {
	params.NSysId = conf.C().OUDNS.SystemID
	params.SProposer = conf.C().OUDNS.Proposer
	params.SOwner = conf.C().OUDNS.Proposer
	resp, err := RequestUdns(conf.C().OUDNS.AddApi, http.MethodPost, params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("AddDomainConfig unmarshal error")
		return nil, err
	}
	key := fmt.Sprintf("udns_ori_%s", params.SDomainName)
	if conf.C().Cache.IsCache { //清除缓存
		if cacheErr := cache.C().Delete(key); cacheErr != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": key,
			}).Error("delete cache error:", cacheErr)
		}
	}
	return
}

// ModDomainConfig 单纯修改域名配置
func (s *service) ModDomainConfig(params *domain.ModDomainParam) (res interface{}, err error) {
	params.NSysId = conf.C().OUDNS.SystemID
	params.SProposer = conf.C().OUDNS.Proposer
	params.SOwner = conf.C().OUDNS.Proposer
	resp, err := RequestUdns(conf.C().OUDNS.ModApi, http.MethodPost, params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"resp":  string(resp),
			"error": err,
		}).Error("ModDomainConfig unmarshal error")
		return nil, err
	}
	key := fmt.Sprintf("udns_ori_%s", params.SDomainName)
	if conf.C().Cache.IsCache { //清除缓存
		if cacheErr := cache.C().Delete(key); cacheErr != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": key,
			}).Error("delete cache error:", cacheErr)
		}
	}
	return
}

// ModifyDomainConfig 校验是否在udns_DB 有域名配置，有修改，无则添加
func (s *service) ModifyDomainConfig(user string, params *domain.ModDomainParam) (res interface{}, err error) {
	if flag := s.CheckUserDomainPermission(user, params.SDomainName); flag == false {
		return res, errors.New("permission deny")
	}
	_, flag, err := s.GetDomainDefaultConfig(params.SDomainName)
	if err != nil {
		s.l.Error("get default config", params.SDomainName, "error", err)
		return
	}
	if !flag {
		res, err = s.AddDomainConfig(params)
	} else {
		res, err = s.ModDomainConfig(params)
	}
	msg, mErr := json.Marshal(params)
	if mErr != nil || err != nil { //转化为字符串
		s.l.Error("#oudns modify api result", res, err)
		s.l.Error("transfer param tp string error: ", mErr)
	}
	newO := domain.RouteConfig{}
	var newDomainList []domain.RRData
	newConfig, nErr := json.Marshal(params.RR) //获取当前RR默认配置
	nErr = json.Unmarshal(newConfig, &newDomainList)
	if nErr != nil {
		s.l.Error("new default config formatter error", nErr)
	}
	for _, nn := range newDomainList {
		if nn.UCountry == 0 && nn.UProvince == 0 && nn.UISP == 0 { //获取默认配置
			newO.SMasterList = s.SplitStr(nn.SMasterList)
			newO.UTTL = nn.UTTL
			newO.Name = params.SDomainName
			break
		}
	}
	l := domain.DomainLog{Description: params.SComment, Operator: user,
		Method: "edit", DomainName: params.SDomainName, Message: string(msg)}
	go s.InsertDomainLog(user, &l)
	return
}

// GetDomainDefaultConfig 获取域名默认配置
func (s *service) GetDomainDefaultConfig(domainName string) (res domain.RouteConfig, flag bool, err error) {
	//flag:true 获取成功，否则获取失败
	flag = true
	q := domain.QueryDomainParam{SDomain: domainName} //从udns获取当前默认配置
	res.Name = domainName
	udnsRes, err := s.GetDomainConfig(&q)
	if err != nil {
		return res, false, err
	}

	curConfig, err := json.Marshal(udnsRes)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"domainName": domainName,
			"error":      err,
		}).Error("get domain Default Conifg formatter error")
		return res, false, err
	}

	var newRes = curConfig
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err = json.Unmarshal(newRes, &o)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"newRes": string(newRes),
			"error":  err,
		}).Error("OudnsResp unmarshal error")
		return res, false, err
	}
	if o.Result != 0 { //获取域名配置失败
		flag = false
	} else {
		m := make(map[string]interface{}) //domain.ModDomainParam{}
		err = json.Unmarshal(curConfig, &m)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"curConfig": string(curConfig),
				"error":     err,
			}).Error("get domain config result error")
			return res, false, err
		}
		var domainList []domain.RRData
		curConfig, err = json.Marshal(m["RR"]) //获取RR配置
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"curConfig": string(curConfig),
				"error":     err,
			}).Error("RRData marshal error")
			return res, false, err
		}
		err = json.Unmarshal(curConfig, &domainList)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"curConfig": string(curConfig),
				"error":     err,
			}).Error("RRData unmarshal error")
			return res, false, err
		}
		for _, dd := range domainList {
			if dd.UCountry == 0 && dd.UProvince == 0 && dd.UISP == 0 {
				res.SMasterList = s.SplitStr(dd.SMasterList)
				res.UTTL = dd.UTTL
				res.Name = domainName
				break
			}
		}
	}
	return res, flag, nil
}

// SyncITApiLog 同步IT接口日志
func (s *service) SyncITApiLog(user string, res domain.ITApiResp, params *domain.ITParam) (err error) {
	success := "是"
	msg := res.Msg
	if res.ExecStatus != 2 {
		success = "否"
	}
	l := domain.SyncApiLog{Operator: user, Success: success, Msg: msg}
	for _, dd := range params.DomainInfo {
		ii, fErr := json.Marshal(&dd)
		if fErr != nil {
			s.l.Error("write IT API Log Error: ", fErr)
		}
		l.DomainName = s.AddPoint(dd.Domain)
		l.Message = string(ii)
		l.Method = dd.OperateType
		l.ApiType = "it"
		err = s.InsertSyncApiLog(user, &l)
		if err != nil {
			s.l.Error("insertSyncApiLog error", err)
		}
	}
	return err
}

// SyncUdnsApiLog 调用udnsWeb和udns_ori Api,unds_openapi操作日志
func (s *service) SyncUdnsApiLog(user, params, domaiName, apiType, method, success, msg string) (err error) {
	l := domain.SyncApiLog{
		ApiType:    apiType,
		DomainName: domaiName,
		Message:    params,
		Method:     method,
		Operator:   user,
		Success:    success,
		Msg:        msg,
	}
	err = s.InsertSyncApiLog(user, &l)
	if err != nil {
		s.l.Error("insert Sync Api Log ERr", err)
	}
	return err
}

// SyncITApi 旧的同步IT接口，更新为udns同步接口
func (s *service) SyncITApi(user string, params *domain.ITParam) (domain.ITApiResp, error) {
	domainInfoArray := params.DomainInfo
	res := domain.ITApiResp{
		Msg:        "",
		Data:       "ok",
		ExecStatus: 2,
	}
	for _, domainInfo := range domainInfoArray {
		switch domainInfo.OperateType {
		case "申请", "变更":
			var udnsParams domain.ModDomainParam
			domainSlices := strings.Split(domainInfo.Domain, ".")
			parentDomain := strings.Join(domainSlices[1:], ".")
			parentDomain = strings.TrimRight(parentDomain, ".") + "."
			q := domain.QueryDomainParam{SDomain: parentDomain}
			parentDomainConfigRes, err := s.GetDomainConfig(&q)
			if err != nil {
				s.l.Error("get domain Config error:" + err.Error())
				return res, err
			}
			parentDomainConfig, err := json.Marshal(parentDomainConfigRes)
			if err != nil {
				s.l.Error("get domain Default Config formatter error:" + err.Error())
				return res, err
			}
			var resData domain.ModDomainParam
			err = json.Unmarshal(parentDomainConfig, &resData)
			if err != nil {
				s.l.Error("json unmarshal domain config error:" + err.Error())
				return res, err
			}
			udnsParams.SDomainName = strings.TrimRight(domainInfo.Domain, ".") + "."
			udnsParams.SISP = "0-0"
			udnsParams.UTTL = domainInfo.TTL
			udnsParams.SCountry = resData.SCountry
			udnsParams.UIDCCount = 1
			udnsParams.URRCount = 1
			udnsParams.UZoneID = resData.UZoneID
			udnsParams.SZoneName = resData.SZoneName
			udnsParams.SArea = "haiwai"
			idc := domain.IdcData{UIdcId: 0, UIdcFlag: 1, SIPList: domainInfo.PServer}
			udnsParams.Idc = append(udnsParams.Idc, idc)
			rr := domain.RRData{SLocalDns: "0.0.0.0", SMasterIDCID: "0", SMasterList: domainInfo.PServer,
				UCountry: 0, UISP: 0, UProvince: 0, UTTL: domainInfo.TTL, UFlag: 0}
			if tool.IsIPv6(domainInfo.PServer) {
				rr.UType = 28
			} else if tool.IsValidDomain(domainInfo.PServer) {
				rr.UType = 5
			} else {
				rr.UType = 1
			}
			udnsParams.RR = append(udnsParams.RR, rr)
			data, _ := json.Marshal(udnsParams)
			s.l.Info("udns request params:" + string(data))
			_, err = s.ModifyDomainConfig(user, &udnsParams)
			if err != nil {
				s.l.Error("udns modify domain config err:" + err.Error())
				return res, err
			}
		case "删除":
			var udnsParams = domain.DelDomainParam{}
			udnsParams.SDomain = strings.TrimRight(domainInfo.Domain, ".") + "."
			udnsParams.SArea = "haiwai"
			_, err := s.DeleteDomainConfig(user, &udnsParams)
			if err != nil {
				s.l.Error("udns delete domain config err:" + err.Error())
				return res, err
			}
		default:
			s.l.Error("param OperateType err:" + domainInfo.OperateType)
			return res, errors.New("param OperateType error")
		}
		return res, nil
	}
	logErr := s.SyncITApiLog(user, res, params)
	if logErr != nil {
		s.l.Error("sync it api error:" + logErr.Error())
	}
	return res, nil
}

// RequestRITApi 请求IT接口
func (s *service) RequestRITApi(domainName string) (res domain.ITRApiResp, err error) {
	q := conf.C().ITApi
	res = domain.ITRApiResp{}
	domainName = s.SplitStr(domainName)
	url := q.RAddr + q.RApi + "?domain_name=" + domainName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return res, err
	}
	//增加token签名
	token := conf.C().ITApi.Token
	timestamp := time.Now().Unix()
	sn := fmt.Sprintf("%v%v%v", timestamp, token, timestamp)
	signature := fmt.Sprintf("%x", sha256.Sum256([]byte(sn)))

	req.Header.Set("signature", signature)
	req.Header.Set("timestamp", fmt.Sprint(timestamp))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	_res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	} else {
		err = json.Unmarshal(_res, &res)
	}
	return
}

// RequestUdns 请求UDNS
func RequestUdns(uri string, method string, body interface{}) (res []byte, err error) {
	q := conf.C().OUDNS
	url := q.Addr + uri
	b, err := json.Marshal(body)
	if err != nil {
		log.C().WithFields(logrus.Fields{
			"url":    url,
			"method": method,
			"body":   body,
			"error":  err,
		}).Error("json marshal error")
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		log.C().WithFields(logrus.Fields{
			"url":    url,
			"method": method,
			"body":   string(b),
			"error":  err,
		}).Error("http NewRequest error")
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.C().WithFields(logrus.Fields{
			"url":    url,
			"method": method,
			"req":    req,
			"error":  err,
		}).Error("request udns api error")
		return nil, err
	}
	defer resp.Body.Close()
	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C().WithFields(logrus.Fields{
			"url":    url,
			"method": method,
			"body":   resp.Body,
			"error":  err,
		}).Error("respond read all error")
		return nil, err
	}
	return
}

// ValidateParam 参数验证
func (s *service) ValidateParam(user, flowType string, param map[string]string) (flag bool, msg string, err error) {
	body := param["body"]
	var domainList []domain.Domain
	err = json.Unmarshal([]byte(body), &domainList)
	if err != nil {
		return false, "", err
	}
	flowApprove := param["flow_approve"]
	switch flowType {
	case flow.ApplyDomainsFlow:
		for _, j := range domainList {
			// 校验域名是否存在
			ok, err := s.IsDomainExisted(j.Name)
			if err != nil {
				return false, "", err
			}
			if ok {
				return false, fmt.Sprintf("%s already existed", j.Name), nil
			}
			// todo 校验默认路由参数合法性，可以晚点再做

		}
	case flow.DeleteDomainsFlow:
		for _, j := range domainList {
			// 校验域名是否存在
			// 同时校验域名是否属于申请人
			d, err := s.DescribeDomain(j.Name)
			if err != nil {
				return false, "", err
			}
			if d.Owner != user {
				return false, fmt.Sprintf("%s is not yours", j.Name), nil
			}
			// todo 这里逻辑还需确认，删除是否直接删除对应域名所有的RR配置，还是说确认没有RR配置之后才能删除？
		}
	case flow.ApplyOwnerPermFlow, flow.ApplyOperatorPermFlow:
		for _, j := range domainList {
			// 校验域名是否存在
			// 同时校验域名都是属于审批步骤人
			d, err := s.DescribeDomain(j.Name)
			if err != nil {
				return false, "", err
			}
			if d.Owner != flowApprove {
				return false, fmt.Sprintf("%s is not %s's domain", j.Name, flowApprove), nil
			}
		}
	case flow.TransferDomainsFlow:
		for _, j := range domainList {
			// 校验域名是否存在
			// 同时校验域名是否属于申请人
			d, err := s.DescribeDomain(j.Name)
			if err != nil {
				return false, "", err
			}
			if d.Owner != user {
				return false, fmt.Sprintf("%s is not yours", j.Name), nil
			}
		}
	default:
		return false, fmt.Sprintf("bad flow type: %s", flowType), nil
	}
	return true, "", nil
}

// SplitStr 字符串分割
func (s *service) SplitStr(ss string) (res string) {
	res = ss
	if len(ss) > 0 {
		i := strings.LastIndex(ss, ".")
		if i == len(ss)-1 {
			curIndex := len(ss) - 1
			res = ss[:curIndex]
		}
	}
	return
}

// AddPoint 字符串后缀补充"."
func (s *service) AddPoint(ss string) (res string) {
	i := strings.LastIndex(ss, ".")
	if i == len(ss)-1 {
		res = ss
	} else {
		res = ss + "."
	}
	return
}

// GenITDomainData 生成IT域名数据
func (s *service) GenITDomainData(
	user string, optType string, new *domain.RouteConfig, old *domain.RouteConfig) (res domain.ITDomainInfo) {
	curDomain := s.SplitStr(new.Name)
	oldValue := s.SplitStr(old.SMasterList)
	newValue := s.SplitStr(new.SMasterList)
	operators := ""
	if new.Operator != "" { //将udns的责任人和权限人结合到一起就是IT的权限人
		curOperator := strings.Split(new.Operator, ";")
		flag := true
		for _, item := range curOperator {
			if item == user {
				flag = false
			}
		}
		if !flag {
			operators = new.Operator
		} else {
			operators = user + ";" + new.Operator
		}
	} else {
		operators = user
	}
	res = domain.ITDomainInfo{
		"",
		curDomain,
		"",
		oldValue,
		false,
		old.UTTL,
		false,
		optType,
		newValue,
		"",
		operators,
		new.UTTL,
		"",
		"",
		2,
	}
	return
}

// GenITApiData 生产IT api请求数据
func (s *service) GenITApiData(user string, d []domain.ITDomainInfo) (res domain.ITParam) {
	res = domain.ITParam{
		user,
		user,
		"",
		"",
		d,
		"",
		"",
		"",
		"",
	}
	return res
}

// genUdnsData 生产UDNS请求数据
func (s *service) genUdnsData(p *domain.RouteConfig) (res domain.ModDomainParam) {
	i := domain.IdcData{SIPList: p.SMasterList, UIdcFlag: 1}
	res.Idc = []domain.IdcData{i}
	r := domain.RRData{
		"0.0.0.0",
		"0",
		p.SMasterList,
		"0",
		"",
		0,
		0,
		0,
		0,
		p.UTTL,
		p.UType}
	res.RR = []domain.RRData{r}
	res.SComment = p.Description
	res.SCountry = "0"
	res.SDomainName = p.Name
	res.SISP = "0-0"
	res.SZoneName = p.Zone
	res.UIDCCount = 1
	res.URRCount = 1
	res.UZoneID = p.ZoneId
	res.SArea = "haiwai"
	res.UTTL = p.UTTL
	return

}

// AddDomainOperation 添加域名操作
func (s *service) AddDomainOperation(user string, j *domain.RouteConfig) (result domain.Message, err error) {
	userInfo, err := s.GetStaffInfo(user)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"error": err,
		}).Error("getUserInfo Error")
	}
	d := s.genUdnsData(j)
	res, err := s.AddDomainConfig(&d) // 调用UDNS后端接口
	if err != nil {
		result.Message = fmt.Sprintf("add Domain config error:%s", err)
		result.Success = false
		return
	}
	newRes, err := json.Marshal(res)
	if err != nil {
		result.Message = fmt.Sprintf("add Domain config error:%s", err)
		result.Success = false
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"res":   res,
			"error": err,
		}).Error("domain config marshal error")
		return
	}
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err = json.Unmarshal(newRes, &o)
	if err != nil {
		result.Message = fmt.Sprintf("add Domain config error:%s", err)
		result.Success = false
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"newRes": newRes,
			"error":  err,
		}).Error("domain config unmarshal error")
		return
	}
	if o.Result != 0 {
		result.Message = o.ErrMsg
		result.Success = false
		return
	} else {
		dd := domain.Domain{
			Name:        j.Name,
			Owner:       user,
			Operator:    j.Operator,
			Description: j.Description,
			DeptID:      int(userInfo["DeptId"].(float64)),
			DeptName:    userInfo["DepartmentName"].(string),
			GroupID:     int(userInfo["GroupId"].(float64)),
			GroupName: fmt.Sprintf("%s-%s", userInfo["DepartmentName"].(string),
				userInfo["GroupName"].(string)),
			BusinessName: j.BusinessName,
		}
		err = s.InsertDomain(user, &dd)
		if err != nil {
			result.Message = fmt.Sprintf("%s, insert domain meta Data error: %s", result.Message, err)
			result.Success = false
			return
		}

		msg, err1 := json.Marshal(d)
		if err1 != nil { //转化为字符串
			s.l.WithFields(logrus.Fields{
				"user":  user,
				"error": err1,
			}).Error("transfer param tp string error")
			result.Message = fmt.Sprintf("transfer param tp string error:%s", err1)
			result.Success = false
			err = errors.New(result.Message)
			return
		}
		l := domain.DomainLog{Description: j.Description, Operator: user,
			Method: "add", DomainName: j.Name, Message: string(msg)}
		s.InsertDomainLog(user, &l) //写域名日志
	}
	return domain.Message{Success: true, Message: "Success"}, nil
}

// DelDomainOperation 删除域名操作
func (s *service) DelDomainOperation(user string, j *domain.RouteConfig) (result domain.Message, err error) {
	q := domain.QueryDomainParam{SDomain: j.Name} //删除前，先获取当前配置
	res, err := s.GetDomainConfig(&q)
	if err != nil {
		result.Message = fmt.Sprintf("get domain Config error:%s", err)
		result.Success = false
		return
	}
	m := make(map[string]interface{}) //domain.ModDomainParam{}
	domainConfig, err := json.Marshal(res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"config": res,
			"error":  err,
		}).Error("domain Config marshal error")
		result.Message = fmt.Sprintf("get domain Config error:%s", err)
		result.Success = false
		return
	}
	err = json.Unmarshal(domainConfig, &m)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"config": string(domainConfig),
			"error":  err,
		}).Error("domain Config unmarshal error")
		result.Message = fmt.Sprintf("get domain Config error:%s", err)
		result.Success = false
		return
	}

	var domainList []domain.RouteConfig
	domainRRConfig, err := json.Marshal(m["RR"]) //获取RR配置
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"config": string(domainConfig),
			"error":  err,
		}).Error("domain Config rr data marshal error")
		result.Message = fmt.Sprintf("get domain rr Config error:%s", err)
		result.Success = false
		return
	}
	err = json.Unmarshal(domainRRConfig, &domainList)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":    user,
			"rr_data": string(domainRRConfig),
			"error":   err,
		}).Error("domain Config rr data unmarshal error")
		result.Message = fmt.Sprintf("get domain rr Config error:%s", err)
		result.Success = false
		return
	}

	d := domain.DelDomainParam{SDomain: j.Name, SArea: "haiwai"}
	deleteRes, err := s.DeleteDomainConfig(user, &d)
	if err != nil {
		result.Message = fmt.Sprintf("delete domain Config error:%s", err)
		result.Success = false
		return
	}
	newRes, err := json.Marshal(deleteRes)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"resp":  deleteRes,
			"error": err,
		}).Error("delete domain config marshal error")
		result.Message = fmt.Sprintf("get domain rr Config error:%s", err)
		result.Success = false
		return
	}
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err = json.Unmarshal(newRes, &o)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":   user,
			"newRes": string(newRes),
			"error":  err,
		}).Error("delete domain config unmarshal error")
		result.Message = fmt.Sprintf("get domain rr Config error:%s", err)
		result.Success = false
		return
	}

	if o.Result != 0 {
		s.l.WithFields(logrus.Fields{
			"user": user,
			"resp": res,
		}).Error("udns delete domain config result")
		result.Message = o.ErrMsg
		result.Success = false
		return
	} else {
		err = s.DeleteDomain(user, j.Name)
		if err != nil {
			result.Message = fmt.Sprintf("%s, delete domain meta Data error: %s", result.Message, err)
			result.Success = false
			return
		}

		l := domain.DomainLog{Description: j.Description, Operator: user,
			Method: "delete", DomainName: j.Name, Message: string(domainConfig)}
		s.InsertDomainLog(user, &l) //写域名日志
	}
	return domain.Message{Success: true, Message: "Success"}, nil
}

// DeliverDomainOperation 修改域名责任人
func (s *service) DeliverDomainOperation(types, user, flowApprove string,
	j *domain.RouteConfig) (result domain.Message, err error) {
	flag := false
	if pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		flag = true
	}
	if s.HasDomainPermission(j.Name, user) {
		flag = true
	}
	if s.HasDomainPermission(j.Name, flowApprove) {
		flag = true
	}
	if !flag {
		result.Message = fmt.Sprintf("permission denied for %s to operate domain: %s", user, j.Name)
		result.Success = false
		s.l.WithFields(logrus.Fields{
			"name":        j.Name,
			"user":        user,
			"types":       types,
			"flowApprove": flowApprove,
		}).Error("permission denied to operate domain")
		err = errors.New("permission denied")
		return
	}
	domainInfo, err := s.DescribeDomain(j.Name)
	if err != nil {
		result.Message = fmt.Sprintf("get domain error: %s", err)
		result.Success = false
		return
	}
	if types == "deliver_domain" {
		domainInfo.Owner = flowApprove
	} else {
		domainInfo.Owner = user
	}
	err = s.UpdateDomain(domainInfo.Owner, domainInfo)
	if err != nil {
		result.Message = fmt.Sprintf("%s, update domain error: %s", result.Message, err)
		result.Success = false
		return
	}

	return domain.Message{Success: true, Message: "Success"}, nil
}

// DomainPermOperation 域名权限判断
func (s *service) DomainPermOperation(user, flowApprove string,
	j *domain.RouteConfig) (result domain.Message, err error) {
	flag := false
	if pkg.Role.HasRolePermission(role.SuperAdminName, flowApprove) {
		flag = true
	}
	if s.HasDomainPermission(j.Name, flowApprove) {
		flag = true
	}
	if s.HasDomainPermission(j.Name, user) {
		flag = true
	}
	if !flag {
		result.Message = fmt.Sprintf("permission denied for %s to operate domain: %s", flowApprove, j.Name)
		result.Success = false
		s.l.WithFields(logrus.Fields{
			"name":        j.Name,
			"user":        user,
			"flowApprove": flowApprove,
		}).Error("permission denied to operate domain")
		err = errors.New("permission denied")
		return
	}
	domainInfo, err := s.DescribeDomain(j.Name)
	if err != nil {
		result.Message = fmt.Sprintf("%s;get domain error: %s", result.Message, err)
		result.Success = false
		return
	}
	curOperator := strings.Split(domainInfo.Operator, ";") //获取当前管理员
	isExist := false
	for _, k := range curOperator {
		if k == user {
			isExist = true
			break
		}
	}
	if !isExist {
		if domainInfo.Operator == "" {
			domainInfo.Operator = user
		} else {
			domainInfo.Operator = fmt.Sprintf("%s;%s", domainInfo.Operator, user)
		}
	} else {
		result.Message = fmt.Sprintf("The user %s already has %s Operator Permission", user, j.Name)
		result.Success = true
		return
	}
	err = s.UpdateDomain(user, domainInfo)
	if err != nil {
		result.Message = fmt.Sprintf("%s, update domain error: %s", result.Message, err)
		result.Success = false
		return
	}
	return domain.Message{Success: true, Message: "Success"}, nil
}

// CheckUserDomainPermission 判断当前用户是否有域名操作权限
func (s *service) CheckUserDomainPermission(user, domainName string) bool {
	flag := false
	// 是否超级管理员
	if pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		flag = true
	}
	// 域名负责人或配置权限人
	if s.HasDomainPermission(domainName, user) {
		flag = true
	}
	return flag
}

// DomainOperation todo 以下所有的操作事项，都需要考虑事务的问题，单个域名操作不成功，需要回退操作
func (s *service) DomainOperation(user, flowType string,
	param map[string]string) (result map[string]domain.Message, err error) {
	var domainList []*domain.RouteConfig
	result = make(map[string]domain.Message)
	err = json.Unmarshal([]byte(param["body"]), &domainList)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":     user,
			"flowType": flowType,
			"param":    param,
			"error":    err,
		}).Error("domain operate param error")
		return
	}
	flowApprove := param["flow_approve"]

	switch flowType {
	case flow.ApplyDomainsFlow, flow.ApplyDomainsFlowCSIG:
		var swg sync.WaitGroup
		for _, j := range domainList {
			swg.Add(1)
			go func(wg *sync.WaitGroup, mark map[string]domain.Message, user string, f *domain.RouteConfig) {
				res, err := s.AddDomainOperation(user, f)
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"user":   user,
						"config": f,
						"error":  err,
					}).Error("domain add error")
				}
				mark[f.Name] = res
				defer wg.Done() //等价于 wg.Add(-1)
			}(&swg, result, user, j)
		}
		swg.Wait()
	case flow.DeleteDomainsFlow:
		var swg sync.WaitGroup
		for _, j := range domainList {
			swg.Add(1)
			go func(wg *sync.WaitGroup, mark map[string]domain.Message, user string, f *domain.RouteConfig) {
				res, err := s.DelDomainOperation(user, f)
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"user":   user,
						"config": f,
						"error":  err,
					}).Error("domain delete error")
				}
				mark[f.Name] = res
				defer wg.Done() //等价于 wg.Add(-1)
			}(&swg, result, user, j)
		}
		swg.Wait()
	case flow.ApplyOperatorPermFlow:
		var swg sync.WaitGroup
		for _, j := range domainList {
			swg.Add(1)
			go func(wg *sync.WaitGroup, mark map[string]domain.Message, user, flowApprove string, f *domain.RouteConfig) {
				res, err := s.DomainPermOperation(user, flowApprove, f) //申请域名管理员权限
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"user":   user,
						"config": f,
						"error":  err,
					}).Error("domain perm operation error")
				}
				mark[f.Name] = res
				defer wg.Done() //等价于 wg.Add(-1)
			}(&swg, result, user, flowApprove, j)
		}
		swg.Wait()
	case flow.ApplyOwnerPermFlow:
		var swg sync.WaitGroup
		for _, j := range domainList {
			swg.Add(1)
			go func(wg *sync.WaitGroup, mark map[string]domain.Message, types, user, flowApprove string, f *domain.RouteConfig) {
				res, err := s.DeliverDomainOperation("apply_owner_perm", user, flowApprove, f) //申请域名责任人权限
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"user":   user,
						"config": f,
						"type":   "apply_owner_perm",
						"error":  err,
					}).Error("deliver domain operation error")
				}
				mark[f.Name] = res
				defer wg.Done()
			}(&swg, result, "apply_domain_perm_owner", user, flowApprove, j)
		}
		swg.Wait()
	case flow.TransferDomainsFlow:
		var swg sync.WaitGroup
		for _, j := range domainList {
			swg.Add(1)
			go func(wg *sync.WaitGroup, mark map[string]domain.Message, types, user, flowApprove string, f *domain.RouteConfig) {
				res, err := s.DeliverDomainOperation("deliver_domain", user, flowApprove, f)
				if err != nil {
					s.l.WithFields(logrus.Fields{
						"user":   user,
						"config": f,
						"type":   "deliver_domain",
						"error":  err,
					}).Error("deliver domain operation error")
				}
				mark[f.Name] = res
				defer wg.Done()
			}(&swg, result, "deliver_domain", user, flowApprove, j)
		}
		swg.Wait()
	default:
		return result, errors.New(fmt.Sprintf("bad flow type: %s", flowType))
	}
	return
}
