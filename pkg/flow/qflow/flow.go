package qflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/pkg/role"
	"github.com/wuzhengjerry/udns-api/tool"
)

// StartTask 启动任务
func (s *service) StartTask(username string, params *flow.StartTaskParams) (res int32, err error) {
	flowMap, err := s.GetFlowMap(username)
	if err != nil {
		return 0, err
	}
	if params.FlowType == flow.ApplyDomainsFlow {
		//判断如果是CSIG用户，则flowType修改为域名申请(CSIG)
		ok, err := tool.IsCSIGUser(username)
		if err != nil {
			s.l.WithFields(logrus.Fields{
				"username": username,
				"error":    err,
			}).Error("tool check csig user error")
			return 0, err
		}
		if ok {
			params.FlowType = flow.ApplyDomainsFlowCSIG
		}
	}
	v, ok := flowMap[params.FlowType]
	if !ok {
		s.l.WithFields(logrus.Fields{
			"flow_type": params.FlowType,
			"map":       flowMap,
		}).Error("can not find flow_type's flow id ")
		return 0, errors.New(fmt.Sprintf(" can's find %s flow id", params.FlowType))
	}
	params.FlowID = v

	// 设置域名审核人
	//if params.FlowType == flow.ApplyDomainsFlow || params.FlowType == flow.DeleteDomainsFlow {
	//	flow_admin, err := pkg.Role.DescribeRole(role.FlowAdminName)
	//	if err != nil {
	//		return 0, err
	//	}
	//	params.EchoStepOwners["operator_approve"] = flow_admin.Users
	//}
	if params.FlowType == flow.ApplyOwnerPermFlow || params.FlowType == flow.ApplyOperatorPermFlow ||
		params.FlowType == flow.TransferDomainsFlow {
		params.EchoStepOwners["flow_approve"] = params.EchoFields["flow_approve"]
	}
	// 发起qflow请求
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/%d/start_by_name",
		conf.C().QFlow.ProjectID, params.FlowID)
	b, err := RequestQFlow(username, uri, http.MethodPost, params)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"params":   params,
			"error":    err,
		}).Error("start by name request qflow error")
		return 0, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return 0, err
	}
	return
}

// DescribeTask 查询任务
func (s *service) DescribeTask(username, taskID string, useCache bool) (task *flow.Task, err error) {
	// 根据参数觉得是否从缓存中获取
	if useCache {
		if conf.C().Cache.IsCache {
			if cache.C().IsExist(flow.TaskCacheKeyPrefix + taskID) {
				if err = cache.C().Get(flow.TaskCacheKeyPrefix+taskID, &task); err != nil {
					s.l.WithFields(logrus.Fields{
						"cache_key": flow.TaskCacheKeyPrefix + taskID,
					}).Error("get cache error:", err)
				} else {
					return
				}
			}
		}
	}
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/%s/detail_task", conf.C().QFlow.ProjectID, taskID)
	b, err := RequestQFlow(username, uri, http.MethodGet, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"taskID":   taskID,
			"error":    err,
		}).Error("detail task request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &task)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}
	// 构建缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(flow.TaskCacheKeyPrefix+taskID, task, role.HalfHourTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": flow.TaskCacheKeyPrefix + taskID,
			}).Error("set cache error:", err)
		}
	}
	return
}

// IsUserTaskCreator 是否申请人
func (s *service) IsUserTaskCreator(username, taskID string) (ok bool, err error) {
	t, err := s.DescribeTask(username, taskID, true)
	if err != nil {
		return
	}
	return t.Creator == username, nil
}

// IsUserStepOwner 是否审批人
func (s *service) IsUserStepOwner(username, taskID, stepID string) (ok bool, err error) {
	t, err := s.DescribeTask(username, taskID, true)
	if err != nil {
		return
	}
	for _, s := range t.Steps {
		if strconv.Itoa(int(s.ID)) == stepID {
			for _, owner := range strings.Split(s.Owner, ";") {
				if owner == username {
					ok = true
				}
			}
			return
		}
	}
	return false, errors.New(fmt.Sprintf("task %s has no step %s", taskID, stepID))
}

// SaveFieldValue 保存审批数据
func (s *service) SaveFieldValue(username, stepID string, params interface{}) (res interface{}, err error) {
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/steps/%s/id_update_step_field",
		conf.C().QFlow.ProjectID, stepID)
	b, err := RequestQFlow(username, uri, http.MethodPut, params)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"params":   params,
			"error":    err,
		}).Error("save field request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}
	return
}

// SubmitStep 提交审批
func (s *service) SubmitStep(username, taskID, stepID string) (res interface{}, err error) {
	// 这里预计需要到qflow查询对应步骤的owner
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/steps/%s/submit_step",
		conf.C().QFlow.ProjectID, stepID)
	b, err := RequestQFlow(username, uri, http.MethodPut, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"taskID":   taskID,
			"stepID":   stepID,
			"error":    err,
		}).Error("submit step request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}

	// 删除缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(flow.TaskCacheKeyPrefix + taskID); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": flow.TaskCacheKeyPrefix + taskID,
			}).Error("delete cache error:", err)
		}
	}
	return res, nil
}

// UpdateStepOwner 修改审批人
func (s *service) UpdateStepOwner(username, taskID, stepID string, owners []string) (res interface{}, err error) {
	//这里做一下特殊判断，限制只能修改直属leader审批这一种步骤
	t, err := s.DescribeTask(username, taskID, true)
	if err != nil {
		return
	}
	ok := false
	for _, s := range t.Steps {
		if strconv.Itoa(int(s.ID)) == stepID && s.Name == flow.LeaderApproveStepName {
			ok = true
		}
	}
	if !ok {
		return nil, errors.New("update step owner only support leader approve step")
	}
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/steps/%s/update_task_step_owner",
		conf.C().QFlow.ProjectID, stepID)
	b, err := RequestQFlow(username, uri, http.MethodPut, owners)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"taskID":   taskID,
			"stepID":   stepID,
			"owners":   owners,
			"error":    err,
		}).Error("update step owner request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}

	// 删除缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(flow.TaskCacheKeyPrefix + taskID); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": flow.TaskCacheKeyPrefix + taskID,
			}).Error("delete cache error:", err)
		}
	}
	// 异步通知用户审批
	go s.NotifyTask(taskID)
	return res, nil
}

// StopTask 终止审批
func (s *service) StopTask(username, taskID string, params interface{}) (res interface{}, err error) {
	// 这里预计需要到qflow查询对应步骤的owner
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/%s/stop", conf.C().QFlow.ProjectID, taskID)
	b, err := RequestQFlow(username, uri, http.MethodPut, params)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"taskID":   taskID,
			"params":   params,
			"error":    err,
		}).Error("stop owner request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}

	// 删除缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().Delete(flow.TaskCacheKeyPrefix + taskID); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": flow.TaskCacheKeyPrefix + taskID,
			}).Error("delete cache error:", err)
		}
	}
	// 异步通知用户审批
	go s.NotifyTask(taskID)
	return res, nil
}

// QueryTask 查询审批任务
func (s *service) QueryTask(username string, params *flow.QueryTaskParams) (res *flow.QueryTaskResp, err error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 10
	}
	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows/tasks/task_list?page=%d&page_size=%d",
		conf.C().QFlow.ProjectID, params.Page, params.PageSize)
	b, err := RequestQFlow(username, uri, http.MethodPost, params)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"params":   params,
			"error":    err,
		}).Error("task list request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}
	return
}

// GetFlowMap 获取流图
func (s *service) GetFlowMap(username string) (flowMap map[string]int32, err error) {
	res, err := s.GetFlowList(username)
	if err != nil {
		return nil, err
	}
	flowMap = make(map[string]int32)
	for _, j := range res.Results {
		flowMap[j.Name] = j.ID
	}
	return
}

// GetFlowList 流图列表
func (s *service) GetFlowList(username string) (res *flow.QueryFlowResp, err error) {
	// 首先从缓存中获取
	if conf.C().Cache.IsCache {
		if cache.C().IsExist(flow.FlowListCacheKey) {
			if err = cache.C().Get(flow.FlowListCacheKey, &res); err != nil {
				s.l.WithFields(logrus.Fields{
					"cache_key": flow.FlowListCacheKey,
				}).Error("get cache error:", err)
			} else {
				return
			}
		}
	}

	uri := fmt.Sprintf("/rest-api/normal/projects/%d/flows", conf.C().QFlow.ProjectID)
	b, err := RequestQFlow(username, uri, http.MethodGet, nil)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"username": username,
			"uri":      uri,
			"error":    err,
		}).Error("flow request qflow error")
		return nil, err
	}
	err = json.Unmarshal(b, &res)
	if err != nil {
		s.l.WithFields(logrus.Fields{
			"res":   b,
			"error": err,
		}).Error("request qflow response error")
		return nil, err
	}
	// 构建缓存
	if conf.C().Cache.IsCache {
		if err = cache.C().PutWithTTL(flow.FlowListCacheKey, res, flow.OneDayTTL); err != nil {
			s.l.WithFields(logrus.Fields{
				"cache_key": flow.FlowListCacheKey,
			}).Error("set cache error:", err)
		}
	}
	return res, nil
}

// TaskAutoOperate 自动实施
func (s *service) TaskAutoOperate(taskID string) (result map[string]domain.Message, err error) {
	t, err := s.DescribeTask(flow.SystemUserName, taskID, true)
	if err != nil {
		return result, err
	}
	param := make(map[string]string)
	for _, j := range *t.Steps[0].Fields {
		param[j.SubmitVarName] = j.InputValue
	}

	result, err = s.d.DomainOperation(t.Creator, t.Flow.Name, param)
	if err != nil {
		return result, err
	}
	return
}

// RequestQFlow 发送qflow请求
func RequestQFlow(username, uri, method string, body interface{}) (res []byte, err error) {
	q := conf.C().QFlow
	url := q.Addr + uri

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+q.Token)
	req.Header.Set("QFLOW-USERNAME", username)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("get business from qlfow error: %s", string(res))
		log.C().WithFields(logrus.Fields{
			"error": errStr,
		}).Error("request qflow error:", errStr)
		return nil, errors.New(errStr)
	}
	return
}
