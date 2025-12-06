// Package http 审批流模块
package http

import (
	"errors"

	"github.com/wuzhengjerry/udns-api/frame/router"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/pkg/route"
)

var (
	api = &handler{}
)

// handler 服务集
type handler struct {
	service flow.Service
}

// Registry 注册HTTP服务路由
func (h *handler) Registry(n *router.Nengine) {
	r := n.Group("/api/v1")
	n.Handle(r, "POST", "flow/task", route.BasePermGroup, false, h.StartTask)
	n.Handle(r, "PUT", "flow/task/:task_id/steps/:step_id/field", route.BasePermGroup,
		false, h.SaveFieldValue)
	n.Handle(r, "PUT", "flow/task/:task_id/steps/:step_id/submit", route.BasePermGroup,
		false, h.SubmitStep)
	n.Handle(r, "PUT", "flow/task/:task_id/steps/:step_id/update_step_owner",
		route.BasePermGroup, false, h.UpdateStepOwner)
	n.Handle(r, "PUT", "flow/task/:task_id/stop", route.BasePermGroup, false, h.StopTask)
	n.Handle(r, "PUT", "flow/task/:task_id/revoke", route.BasePermGroup, false, h.RevokeTask)
	// 查询接口
	n.Handle(r, "GET", "flow/task/:task_id", route.BasePermGroup, false, h.DescribeTask)
	n.Handle(r, "POST", "flow/tasks", route.BasePermGroup, false, h.AllTaskList)
	n.Handle(r, "POST", "flow/tasks/my_application", route.BasePermGroup, false, h.MyApplication)
	n.Handle(r, "POST", "flow/tasks/my_approve", route.BasePermGroup, false, h.MyApprove)
	n.Handle(r, "GET", "flows", route.BasePermGroup, false, h.AllFlowList)
	// qflow 交互
	n.Handle(r, "GET", "flow/status_notify", route.BasePermGroup, false, h.StatusNotify)
	n.Handle(r, "POST", "flow/task_auto_operate", route.BasePermGroup, false, h.TaskAutoOperate)
}

// Config 配置
func (h *handler) Config() error {
	if pkg.Flow == nil {
		return errors.New("dependence application module Flow is nil")
	}

	h.service = pkg.Flow
	return nil
}

// init 初始化
func init() {
	pkg.RegistryHTTPV1("flow", api)
}
