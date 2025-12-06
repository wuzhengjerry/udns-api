package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/openapi"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ValidateQueryParam 校验查询参数合法性
func (s *service) ValidateQueryParam(p *openapi.QueryParam) (res bson.M, err error) {
	filter := bson.M{"deleted": false}
	if len(p.Owner) > 0 {
		filter["owner"] = bson.M{"$in": p.Owner}
	}
	if len(p.GroupID) > 0 {
		filter["group_id"] = bson.M{"$in": p.GroupID}
	}
	if len(p.DeptID) > 0 {
		filter["dept_id"] = bson.M{"$in": p.DeptID}
	}
	if p.Description != "" {
		filter["description"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + p.Description + ".*"}}
	}
	var oriList []bson.M
	var curList []bson.M
	var optList []bson.M
	if len(p.Domains) > 0 {
		for _, item := range p.Domains {
			tmp := item
			if string(item[0]) == "*" {
				tmp = item[1:]
			}
			tmp = strings.ReplaceAll(tmp, "*", "\\*")
			tmp = strings.ReplaceAll(tmp, ".", "\\.")
			oriList = append(oriList,
				bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: ".*" + tmp + ".*", Options: "i"}}})
		}
		filter["$or"] = oriList
	} else {
		err = errors.New("user has no domains")
		s.l.WithFields(logrus.Fields{
			"user":  p.User,
			"error": err,
		}).Error("user has no domains")
		return
	}
	if len(p.Name) > 0 {
		flag, err1 := s.isDomainPerm(p.Domains, p.Name)
		if err1 != nil {
			s.l.WithFields(logrus.Fields{
				"user":    p.User,
				"Name":    p.Name,
				"Domains": p.Domains,
				"error":   err1,
			}).Error("domain permission deny")
			err = err1
			return
		}
		if !flag {
			err = errors.New("domains are not both allowed")
			return
		} else {
			filter["name"] = bson.M{"$in": p.Name}
		}
	}
	if len(p.Regulator) > 0 {
		if len(p.Owner) > 0 || len(p.Operator) > 0 {
			err = errors.New("Invalid Param: Regulator and Owner or Operator can't both exit ")
			return
		}
		for _, item := range p.Regulator {
			curList = append(
				curList,
				bson.M{"operator": bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + item + "(;|$)"}}},
				bson.M{"owner": item})
		}
		filter["$and"] = []bson.M{bson.M{"$or": oriList}, bson.M{"$or": curList}}
	}
	if len(p.Operator) > 0 {
		for _, item := range p.Operator {
			optList = append(optList,
				bson.M{"operator": bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + item + "(;|$)"}}})
		}
		filter["$and"] = []bson.M{bson.M{"$or": oriList}, bson.M{"$or": optList}}
	}
	if len(oriList) > 0 && len(curList) == 0 && len(optList) == 0 {
		filter["$or"] = oriList
	}
	res = filter
	return
}

// ValidateGlobalQueryParam 验证参数是否合法
func (s *service) ValidateGlobalQueryParam(p *openapi.QueryParam) (res bson.M, err error) {
	filter := bson.M{"deleted": false}
	if len(p.Owner) > 0 {
		filter["owner"] = bson.M{"$in": p.Owner}
	}
	if len(p.GroupID) > 0 {
		filter["group_id"] = bson.M{"$in": p.GroupID}
	}
	if len(p.DeptID) > 0 {
		filter["dept_id"] = bson.M{"$in": p.DeptID}
	}
	if p.Description != "" {
		filter["description"] = bson.M{"$regex": primitive.Regex{Pattern: ".*" + p.Description + ".*"}}
	}
	var curList []bson.M
	var optList []bson.M
	if len(p.Name) > 0 {
		filter["name"] = bson.M{"$in": p.Name}
	}
	if len(p.Regulator) > 0 {
		if len(p.Owner) > 0 || len(p.Operator) > 0 {
			err = errors.New("Invalid Param: Regulator and Owner or Operator can't both exit ")
			return filter, err
		}
		for _, item := range p.Regulator {
			curList = append(
				curList,
				bson.M{"operator": bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + item + "(;|$)"}}},
				bson.M{"owner": item})
		}
		filter["or"] = curList
	}
	if len(p.Operator) > 0 {
		for _, item := range p.Operator {
			optList = append(optList,
				bson.M{"operator": bson.M{"$regex": primitive.Regex{Pattern: "(;|^)" + item + "(;|$)"}}})
		}
		filter["or"] = curList
	}
	res = filter
	return
}

// QueryDomain 查询域名
func (s *service) QueryDomain(p *openapi.QueryParam) (count int64, rs []*openapi.DomainInfo, err error) {
	limit := p.Limit
	if limit > 1000 {
		err = errors.New("单次请求最大值为1000条")
		limit = 1000
		//return 0, rs, err
	}
	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(p.Offset)
	findOptions.SetSort(bson.M{"created_time": -1})
	filter, filterErr := s.ValidateGlobalQueryParam(p)
	if filterErr != nil {
		s.l.WithFields(logrus.Fields{
			"params": p,
			"error":  filterErr,
		}).Error("ValidateGlobalQueryParam failed")
		return 0, rs, filterErr
	}
	var res []*domain.Domain
	count, err = s.col.CountDocuments(context.TODO(), filter)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"filter": filter,
			"error":  err,
		}).Error("open api query domain count error")
		return count, rs, err
	}
	cur, err := s.col.Find(context.TODO(), filter, findOptions)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"filter": filter,
			"error":  err,
		}).Error("open api query domain error")
		return count, rs, err
	}
	for cur.Next(context.TODO()) {
		var d domain.Domain
		err = cur.Decode(&d)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"error": err,
			}).Error("domain decode error")
			return count, rs, err
		}
		res = append(res, &d)
	}
	newConfig, nErr := json.Marshal(res)
	nErr = json.Unmarshal(newConfig, &rs)
	if nErr != nil {
		s.l.WithFields(logrus.Fields{
			"res":   res,
			"error": nErr,
		}).Error("domain config marsha error")
		return count, rs, nErr
	}
	work := openapi.NewPool(openapi.QueryWorkCount)
	for _, j := range rs {
		work.Add(1)
		q := domain.QueryDomainParam{SDomain: j.Name}
		go func(wg *openapi.WaitGroup,
			mark []*openapi.DomainInfo, info *openapi.DomainInfo, params domain.QueryDomainParam) {
			domainInfo, err1 := s.GetDomainConfigData(info, params)
			if err1 != nil {
				s.l.WithFields(logrus.Fields{
					"info":   info,
					"params": params,
					"error":  err1,
				}).Error("getDomainConfigData error")
			}
			mark = append(mark, domainInfo)
			defer wg.Done()
		}(work, rs, j, q)
	}
	work.Wait()
	return count, rs, nil
}

// GetDomainConfigData 获取域名配置
func (s *service) GetDomainConfigData(d *openapi.DomainInfo,
	p domain.QueryDomainParam) (res *openapi.DomainInfo, err error) {
	udnsRes, err := pkg.Domain.GetDomainConfig(&p)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"params": p,
			"error":  err,
		}).Error("getDomainConfigData error")
		return
	}
	curConfig, err := json.Marshal(udnsRes)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"udnsRes": udnsRes,
			"error":   err,
		}).Error("getDomainConfigData error")
		return
	}
	var newRes = curConfig
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err = json.Unmarshal(newRes, &o)
	if o.Result != 0 {
		err = errors.New(o.ErrMsg)
		return
	} else {
		m := make(map[string]interface{}) //domain.ModDomainParam{}
		err = json.Unmarshal(curConfig, &m)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unmarshal domain config result error")
		}
		var domainList []domain.RRData
		curConfig, _ = json.Marshal(m["RR"])
		err = json.Unmarshal(curConfig, &domainList)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unmarshal curConfig config result error")
			return
		}
		var config []*openapi.DomainConfig
		for _, item := range domainList {
			tmp := openapi.DomainConfig{}
			route := fmt.Sprintf("%d-%d-%d", item.UCountry, item.UProvince, item.UISP)
			tmp.Route = route
			tmp.Content = item.SMasterList
			tmp.UType = item.UType
			config = append(config, &tmp)
		}
		ttl, _ := m["uTTL"].(int32)
		d.UTTL = ttl
		d.Config = config
	}
	res = d
	return
}

/*
  - isDomainPerm 域名权限校验
    domains 用于配置域名权限
    names 请求传入域名
*/

func (s *service) isDomainPerm(domains []string, names []string) (flag bool, err error) {
	result := 0
	if len(domains) <= 0 {
		err = errors.New("no Domain Operation perm")
		return false, err
	}
	for _, info := range names {
		curFlag := false

		for _, item := range domains {
			if len(item) > 0 {
				if string(item[0]) == "*" {
					tmp := item[1:]
					flags := s.isEnd(info, tmp)
					if flags {
						curFlag = true
						break
					}

					if item[2:] == info {
						curFlag = true
						break
					}
				}
			}
		}
		if curFlag {
			result += 1
		}
	}
	if result == len(names) {
		flag = true
	} else {
		flag = false
	}
	return
}

// isEnd 是否后缀结尾
func (s *service) isEnd(obj string, opt string) (flag bool) {
	return len(obj) >= len(opt) && obj[len(obj)-len(opt):] == opt
}

// TransferData 域名转移
func (s *service) TransferData(d []*domain.ModDomainParam,
	o []*openapi.DomainInfo) (res []*openapi.DomainInfo, err error) {
	for i, info := range o {
		for _, jInfo := range d {
			if info.Name == jInfo.SDomainName {
				o[i].UTTL = jInfo.UTTL
			}
			var config []*openapi.DomainConfig
			for _, item := range jInfo.RR {
				tmp := openapi.DomainConfig{}
				route := fmt.Sprintf("%d-%d-%d", item.UCountry, item.UISP, item.UProvince)
				tmp.Route = route
				tmp.Content = item.SMasterList
				tmp.UType = item.UType
				config = append(config, &tmp)
			}
			o[i].Config = config
		}
	}
	res = o
	return
}

// AddDomain 新增域名
func (s *service) AddDomain(d []*openapi.DomainInfo, user string, domains []string) (final interface{}, err error) {
	var names []string
	for _, item := range d {
		curFlag, iErr := s.IsParamValidate(item)
		if iErr != nil {
			s.l.WithFields(logrus.Fields{
				"error": iErr,
			}).Error("parameter verification failure")
			return final, iErr
		}
		if !curFlag {
			err = errors.New("parameter verification failure")
			return
		}
		tmp := item.Name
		if len(tmp) == 0 { //域名为空可直接返回异常
			err = errors.New("parameter name null")
			return
		}
		//验证是否以.结尾,没有则要加"点"，默认DB存储域名加点
		if tmp[len(tmp)-1:] != "." {
			tmp = strings.TrimSpace(tmp) + "."
		}
		item.Name = tmp
		names = append(names, tmp)
	}
	flag, err := s.isDomainPerm(domains, names)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":    user,
			"names":   names,
			"domains": domains,
			"error":   err,
		}).Error("domain permission deny")
		return
	}
	if !flag {
		err = errors.New("exist Unauthorized Domain")
		return
	}
	result := make(map[string]domain.Message)
	num := len(names)
	if num > openapi.OptWorkCount {
		num = openapi.OptWorkCount
	}
	work := openapi.NewPool(num)
	for _, j := range d {
		work.Add(1)
		go func(wg *openapi.WaitGroup, mark map[string]domain.Message, user string, f *openapi.DomainInfo) {
			res, err1 := s.AsyncOptDomainOperation(user, "申请", f)
			if err1 != nil {
				s.l.WithFields(logrus.Fields{
					"user":   user,
					"name":   f.Name,
					"opType": "申请",
					"error":  err1,
				}).Error("AsyncOptDomainOperation error")
			}
			mark[f.Name] = res
			defer wg.Done()
		}(work, result, user, j)
	}
	work.Wait()
	msg := ""
	for _, curRes := range result {
		if !curRes.Success {
			msg += curRes.Message + ";"
		}
	}
	if msg != "" {
		err = errors.New(msg)
	}
	return result, err
}

// ModDomain 修改域名
func (s *service) ModDomain(d []*openapi.DomainInfo, user string, domains []string) (final interface{}, err error) {
	var names []string
	for _, item := range d {
		curFlag, itErr := s.IsParamValidate(item)
		if itErr != nil {
			s.l.WithFields(logrus.Fields{
				"user":    user,
				"domains": domains,
				"error":   itErr,
			}).Error("param validate error")
			return final, itErr
		}
		if !curFlag {
			err = errors.New("parameter verification failure")
			return
		}
		tmp := item.Name
		if len(tmp) == 0 {
			err = errors.New("parameter verification failure")
			return
		}
		if tmp[len(tmp)-1:] != "." {
			tmp = strings.TrimSpace(tmp) + "."
		}
		item.Name = tmp
		names = append(names, tmp)
	}
	flag, err := s.isDomainPerm(domains, names)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":    user,
			"names":   names,
			"domains": domains,
			"error":   err,
		}).Error("domain permission deny")
		return
	}
	if !flag {
		err = errors.New("exist Unauthorized Domain")
		return
	}
	result := make(map[string]domain.Message)
	num := len(d)
	if num > openapi.OptWorkCount {
		num = openapi.OptWorkCount
	}
	work := openapi.NewPool(num)
	for _, j := range d {
		work.Add(1)
		go func(wg *openapi.WaitGroup, mark map[string]domain.Message, user string, f *openapi.DomainInfo) {
			res, err1 := s.AsyncOptDomainOperation(user, "变更", f)
			if err1 != nil {
				s.l.WithFields(logrus.Fields{
					"user":   user,
					"name":   f.Name,
					"opType": "变更",
					"error":  err1,
				}).Error("AsyncOptDomainOperation error")
			}
			mark[f.Name] = res
			defer wg.Done() //等价于 wg.Add(-1)
		}(work, result, user, j)
	}
	work.Wait()
	msg := ""
	for _, curRes := range result {
		if !curRes.Success {
			msg += curRes.Message + ";"
		}
	}
	if msg != "" {
		err = errors.New(msg)
	}
	return result, err
}

// DelDomain 删除域名
func (s *service) DelDomain(d *openapi.DelDomainParam, user string, domains []string) (final interface{}, err error) {
	var names []string
	if len(d.Name) == 0 || d.Reason == "" {
		err = errors.New("parameter verification failure")
		return
	}
	for i, item := range d.Name {
		tmp := item
		if len(tmp) == 0 { //域名为空可直接返回异常
			err = errors.New("parameter name null error")
			return
		}
		//验证是否以.结尾,没有则要加"点"，默认DB存储域名加点
		if tmp[len(tmp)-1:] != "." {
			tmp = strings.TrimSpace(tmp) + "."
		}
		d.Name[i] = tmp
		names = append(names, tmp)
	}
	flag, err := s.isDomainPerm(domains, names) //检查是否为有权限操作的域名
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"user":    user,
			"names":   names,
			"domains": domains,
			"error":   err,
		}).Error("domain permission deny")
		return
	}
	if !flag {
		err = errors.New("exist Unauthorized Domain")
		return
	}
	result := make(map[string]domain.Message)
	num := len(names)
	if num > openapi.OptWorkCount {
		num = openapi.OptWorkCount
	}
	work := openapi.NewPool(num)
	for _, j := range d.Name {
		work.Add(1)
		go func(wg *openapi.WaitGroup, mark map[string]domain.Message, user, domainName, reason string) {
			res, err1 := s.AsyncDeleteDomain(user, domainName, reason)
			if err1 != nil {
				s.l.WithFields(logrus.Fields{
					"user":       user,
					"domainName": domainName,
					"reason":     reason,
					"error":      err1,
				}).Error("AsyncDeleteDomain error")
			}
			mark[domainName] = res
			defer wg.Done() //等价于 wg.Add(-1)
		}(work, result, user, j, d.Reason)
	}
	work.Wait()
	msg := ""
	for _, curRes := range result {
		if !curRes.Success {
			msg += curRes.Message + ";"
		}
	}
	if msg != "" {
		err = errors.New(msg)
	}
	return result, err
}

// AsyncDeleteDomain 删除域名
func (s *service) AsyncDeleteDomain(user, domainName, reason string) (result domain.Message, err error) {
	q := domain.QueryDomainParam{SDomain: domainName} //删除前，先获取当前配置
	res, mErr := pkg.Domain.GetDomainConfig(&q)
	if mErr != nil {
		s.l.WithFields(logrus.Fields{
			"user":       user,
			"domainName": domainName,
			"reason":     reason,
			"error":      mErr,
		}).Error("get_domain_config error")
		result.Message = fmt.Sprintf("get domain Config error:%s", mErr)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	curConfig, err1 := json.Marshal(res)
	if err1 != nil {
		s.l.Error("domain config marshal error:", err1)
		result.Message = fmt.Sprintf("domain Config marshal error:%s", err1)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	m := make(map[string]interface{})
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err2 := json.Unmarshal(curConfig, &o)
	if err2 != nil {
		s.l.Error("curConfig unmarshal to o error:", err2)
		result.Message = fmt.Sprintf("domain Config unmarshal error:%s", err2)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	if o.Result != 0 { //获取配置失败直接返回，域名很可能不存在
		result.Message = o.ErrMsg
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	err3 := json.Unmarshal(curConfig, &m)
	if err3 != nil {
		if err3 != nil {
			s.l.Error("curConfig unmarshal to m error:", err3)
			result.Message = fmt.Sprintf("domain Config unmarshal error:%s", err3)
			result.Success = false
			err = errors.New(result.Message)
			return
		}
	}
	var domainList []*domain.RouteConfig
	value, ok := m["RR"]
	if !ok {
		result.Message = fmt.Sprintf("get %s RR Config Error", domainName)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	curConfig, _ = json.Marshal(value) //获取RR配置
	err4 := json.Unmarshal(curConfig, &domainList)
	if err4 != nil {
		s.l.Error("curConfig unmarshal to domainList error:", err4)
		result.Message = fmt.Sprintf("domain Config unmarshal error:%s", err4)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	msg, mErr := json.Marshal(res)
	if mErr != nil {
		result.Message = fmt.Sprintf("transfer param tp string error:%s", mErr)
		result.Success = false
		s.l.Error("transfer param tp string error:", mErr)
		err = errors.New(result.Message)
		return
	}
	d := domain.DelDomainParam{SDomain: domainName, SArea: "haiwai"}
	dRes, dErr := pkg.Domain.DeleteDomainConfigAction(&d) // 调用UDNS后端接口
	if dErr != nil {
		s.l.WithFields(logrus.Fields{
			"user":  user,
			"param": d,
			"error": dErr,
		}).Error("DeleteDomainConfigAction error")
		result.Message = fmt.Sprintf("delete domain config error:%s", mErr)
		result.Success = false
		err = errors.New(result.Message)
		return
	}
	newRes, _ := json.Marshal(dRes)
	err5 := json.Unmarshal(newRes, &o)
	if err5 != nil {
		s.l.Error("newRes unmarshal to o error:", err5)
		result.Message = fmt.Sprintf("delete domain config res error:%s", err5)
		result.Success = false
		err = errors.New(result.Message)
		return
	}

	if o.Result != 0 {
		s.l.Error("o.ErrMsg:", o.ErrMsg)
		result.Message = o.ErrMsg
		result.Success = false
		err = errors.New(result.Message)
		return
	} else {
		l := domain.DomainLog{Description: reason, Operator: user,
			Method: "delete", DomainName: domainName, Message: string(msg)}
		go func(user string, l *domain.DomainLog) {
			logErr := pkg.Domain.InsertDomainLog(user, l)
			if logErr != nil {
				s.l.WithFields(logrus.Fields{
					"user":  user,
					"value": l,
					"error": logErr,
				}).Error("Insert Domain Log error")
			}
		}(user, &l)
		domainErr := pkg.Domain.DeleteDomain(user, domainName)
		if domainErr != nil {
			result.Message = fmt.Sprintf("%s, insert domain meta Data error: %s", result.Message, dErr)
			result.Success = false
			err = errors.New(result.Message)
			s.l.WithFields(logrus.Fields{
				"user":       user,
				"domainName": domainName,
				"error":      domainErr,
			}).Error("deleteDomain error")
			return
		}
	}
	return domain.Message{Success: true, Message: "Success"}, err
}

// IsUserValidate 用户鉴权
func (s *service) IsUserValidate(name string) (flag bool, domains []string, err error) {
	res, userErr := pkg.WebUser.DescribeUser(name)
	if userErr != nil {
		s.l.WithFields(logrus.Fields{
			"name":  name,
			"error": userErr,
		}).Error("check user validate error")
		return false, domains, userErr
	}
	domains = res.Domain
	return true, domains, nil
}

// IsParamValidate 参数校验
func (s *service) IsParamValidate(p *openapi.DomainInfo) (flag bool, err error) {
	if p.Name != "" && p.UTTL != 0 && p.Owner != "" && len(p.Config) > 0 {
		return true, nil
	} else {
		return false, errors.New("parameter verification failure,name,uTTL,owner or config is idle")
	}
}

// ValidateDomainContent 校验域名内容
func (s *service) ValidateDomainContent(item *openapi.DomainConfig) (flag bool, err error) {
	ipReg := `^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$`
	//domainReg := `^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?$`
	// domainReg := `([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}`
	domainReg := `[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?$`
	aaaaReg := `^([\da-fA-F]{1,4}:){6}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)|::(
[\da−fA−F]1,4:)0,4((25[0−5]|2[0−4]\d|[01]?\d\d?)\.)3(25[0−5]|2[0−4]\d|[01]?\d\d?)|::([\da−fA−F]1,4:)0,
4((25[0−5]|2[0−4]\d|[01]?\d\d?)\.)3(25[0−5]|2[0−4]\d|[01]?\d\d?)|^([\da-fA-F]{1,4}:):([\da-fA-F]{1,4}:){0,3}
((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)|([\da−fA−F]1,4:)2:([\da−fA−F]1,4:)0,2((25[0−5]
|2[0−4]\d|[01]?\d\d?)\.)3(25[0−5]|2[0−4]\d|[01]?\d\d?)|([\da−fA−F]1,4:)2:([\da−fA−F]1,4:)0,2((25[0−5]|2[0−4]\d
|[01]?\d\d?)\.)3(25[0−5]|2[0−4]\d|[01]?\d\d?)|^([\da-fA-F]{1,4}:){3}:([\da-fA-F]{1,4}:){0,1}((25[0-5]|2[0-4]\d
|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)|([\da−fA−F]1,4:)4:((25[0−5]|2[0−4]\d|[01]?\d\d?)\.)3(25[0−5]
|2[0−4]\d|[01]?\d\d?)|([\da−fA−F]1,4:)4:((25[0−5]|2[0−4]\d|[01]?\d\d?)\.)3(25[0−5]|2[0−4]\d|[01]?\d\d?)
|^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}|:((:[\da−fA−F]1,4)1,6|:)|:((:[\da−fA−F]1,4)1,6|:)|^[\da-fA-F]
{1,4}:((:[\da-fA-F]{1,4}){1,5}|:)|([\da−fA−F]1,4:)2((:[\da−fA−F]1,4)1,4|:)|([\da−fA−F]1,4:)2((:[\da−fA−F]1,4)1,4|:)
|^([\da-fA-F]{1,4}:){3}((:[\da-fA-F]{1,4}){1,3}|:)|([\da−fA−F]1,4:)4((:[\da−fA−F]1,4)1,2|:)
|([\da−fA−F]1,4:)4((:[\da−fA−F]1,4)1,2|:)|^([\da-fA-F]{1,4}:){5}:([\da-fA-F]{1,4})?
|([\da−fA−F]1,4:)6:|([\da−fA−F]1,4:)6:`
	contents := strings.Split(item.Content, ";")
	switch item.UType {
	case 1:
		r, _ := regexp.Compile(ipReg)
		if r == nil { //判断是否为空
			msg := fmt.Sprintf("%s is invalid", item.Content)
			return false, errors.New(msg)
		}
		for _, content := range contents {
			if !r.MatchString(content) {
				msg := fmt.Sprintf("DNS content [%s] is invalid， when type is %d ", content, item.UType)
				err = errors.New(msg)
				return false, err
			}
		}

	case 5:
		r, _ := regexp.Compile(domainReg)
		if r == nil { //判断是否为空
			msg := fmt.Sprintf("%s is invalid", item.Content)
			return false, errors.New(msg)
		}
		ipMatch, _ := regexp.Compile(ipReg)
		for _, content := range contents {
			if !r.MatchString(content) || ipMatch.MatchString(content) { //是域名且非Ip地址
				msg := fmt.Sprintf("DNS content [%s] is invalid， when type is %d ", content, item.UType)
				err = errors.New(msg)
				return false, err
			}
		}
	case 28:
		r, _ := regexp.Compile(aaaaReg)
		if r == nil { //判断是否为空
			msg := fmt.Sprintf("%s is invalid", item.Content)
			return false, errors.New(msg)
		}
		for _, content := range contents {
			if !r.MatchString(content) {
				err = errors.New(
					fmt.Sprintf("DNS content [%s] is invalid， when type is %d ", content, item.UType))
				return false, err
			}
		}
	default:
		err = errors.New(fmt.Sprintf("DNS type is invalid: %d", item.UType))
		return false, err
	}
	return true, err

}

// GenUdnsData 生产UDNS数据
func (s *service) GenUdnsData(p *openapi.DomainInfo) (res domain.ModDomainParam, err error) {
	var country []string
	var isp []string
	var idcs []domain.IdcData
	var rrs []domain.RRData
	curZone := "woa.com."
	z, zoneErr := pkg.Zone.DescribeZone(curZone)
	if zoneErr != nil {
		s.l.WithFields(logrus.Fields{
			"zone":  curZone,
			"error": zoneErr,
		}).Error("describe zone error")
		return res, zoneErr
	}
	zoneId := int32(z.ZoneID)
	for i, item := range p.Config {
		ok, routeErr := pkg.Route.IsRouteExisted(item.Route)
		if routeErr != nil {
			s.l.WithFields(logrus.Fields{
				"route": item.Route,
				"error": routeErr,
			}).Error("query route error")
		}
		if !ok {
			msg := fmt.Sprintf("%s is not Exist", item.Route)
			err = errors.New(msg)
			return
		}
		route := strings.Split(item.Route, "-")
		if len(route) < 3 {
			msg := fmt.Sprintf("%s is invalid", item.Route)
			err = errors.New(msg)
			return
		}
		cc, _ := strconv.Atoi(route[0])
		pp, _ := strconv.Atoi(route[1])
		ii, _ := strconv.Atoi(route[2])

		if !contains(country, route[0]) {
			country = append(country, route[0])
		}

		genIsp := fmt.Sprintf("%v-%v", ii, pp)
		if !contains(isp, genIsp) {
			isp = append(isp, genIsp)
		}

		if item.Content == "" {
			msg := fmt.Sprintf("DNS content is idle: %s ", item.Content)
			err = errors.New(msg)
			return
		} else {
			flag, validateErr := s.ValidateDomainContent(item)
			if !flag || validateErr != nil {
				s.l.WithFields(logrus.Fields{
					"item":  item,
					"error": validateErr,
				}).Error("domain content validate error")
				return res, validateErr
			}
		}
		idc := domain.IdcData{SIPList: item.Content, UIdcFlag: 1, UIdcId: int32(i)}
		tContent := item.Content
		if item.UType == 5 { //解析域名要加点
			if tContent[len(tContent)-1:] != "." {
				tContent += "."
			}
		}
		rr := domain.RRData{SLocalDns: "0.0.0.0", SMasterIDCID: fmt.Sprintf("%d", i), SMasterList: tContent,
			SSlaveIDCID: "0", UCountry: int32(cc), UISP: int32(ii),
			UProvince: int32(pp), UTTL: p.UTTL, UType: item.UType}
		rrs = append(rrs, rr)
		idcs = append(idcs, idc)
	}
	idcCount := len(idcs)
	rrCount := len(rrs)
	res.RR = rrs
	res.Idc = idcs
	res.SComment = p.Description
	res.SCountry = strings.Join(country, ";")
	res.SDomainName = p.Name
	res.SISP = strings.Join(isp, ";")
	res.SZoneName = curZone
	res.UIDCCount = int32(idcCount)
	res.URRCount = int32(rrCount)
	res.UZoneID = zoneId
	res.SArea = "haiwai"
	res.UTTL = p.UTTL
	return
}

// AsyncOptDomainOperation 新增或修改操作
func (s *service) AsyncOptDomainOperation(user string, optType string,
	j *openapi.DomainInfo) (result domain.Message, err error) {
	userInfo, e := pkg.Domain.GetStaffInfo(j.Owner)
	if e != nil {
		s.l.WithFields(logrus.Fields{
			"user":    user,
			"optType": optType,
			"error":   e,
		}).Error("get staff info error")
		result = domain.Message{Success: false, Message: fmt.Sprintf("get owner info error:%s", e)}
		err = errors.New(result.Message)
		return
	}
	d, pErr := s.GenUdnsData(j)
	if pErr != nil {
		return domain.Message{Message: fmt.Sprintf("%s", pErr), Success: false}, pErr
	}
	var res interface{}
	if optType == "申请" {
		res, err = pkg.Domain.AddDomainConfig(&d)
	} else {
		res, err = pkg.Domain.ModDomainConfig(&d)
	}
	if err != nil {
		return domain.Message{Success: false, Message: fmt.Sprintf("opt domain config error:%s", err)}, err
	}
	newRes, err := json.Marshal(res)
	o := domain.OudnsResp{ErrMsg: "", Result: 0}
	err = json.Unmarshal(newRes, &o)
	if o.Result != 0 {
		return domain.Message{Success: false, Message: o.ErrMsg}, errors.New(o.ErrMsg)
	} else {
		msg, mErr := json.Marshal(d)
		if mErr != nil {
			result = domain.Message{Success: false, Message: fmt.Sprintf("transfer param format error:%s", mErr)}
			err = errors.New(result.Message)
			return
		}
		var groupId, deptId int
		if userInfo["GroupId"] != nil {
			groupId = int(userInfo["GroupId"].(float64))
		}
		if userInfo["DeptId"] != nil {
			deptId = int(userInfo["DeptId"].(float64))
		}
		dd := domain.Domain{
			Name: j.Name, Owner: j.Owner, Operator: j.Operator, Description: j.Description,
			DeptID: deptId, DeptName: userInfo["DepartmentName"].(string), GroupID: groupId,
			GroupName: fmt.Sprintf("%s-%s", userInfo["DepartmentName"].(string),
				userInfo["GroupName"].(string)), BusinessName: j.BusinessName,
		}
		method := "add"
		if optType == "申请" {
			dErr := pkg.Domain.InsertDomain(j.Owner, &dd)
			if dErr != nil {
				result.Message = fmt.Sprintf("%s, add domain meta error: %s", result.Message, dErr)
				result.Success = false
				err = errors.New(result.Message)
				return
			}
		} else {
			method = "edit"
			dErr := pkg.Domain.UpdateDomain(user, &dd)
			if dErr != nil {
				result.Success = false
				result.Message = fmt.Sprintf("%s, update domain meta error: %s", result.Message, dErr)
				err = errors.New(result.Message)
				return
			}
		}

		l := domain.DomainLog{Description: j.Description, Operator: user,
			Method: method, DomainName: j.Name, Message: string(msg)}
		go func(user string, l *domain.DomainLog) {
			errN := pkg.Domain.InsertDomainLog(user, l)
			if errN != nil {
				s.l.WithFields(logrus.Fields{
					"user":  user,
					"value": l,
					"error": errN,
				}).Error("insert domain log error")
			}
		}(user, &l)
	}
	return domain.Message{Success: true, Message: "Success"}, nil
}

// contains str是否在数组s中
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
