package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	config "github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/frame"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/flow"
	"github.com/wuzhengjerry/udns-api/pkg/role"
)

// StartTask 启动任务
// @Summary 新增流程任务
// @version 1.0.0
// @tags 流程管理
// @description  传递参数新增task
// @Produce json
// @Param StaffName header string true "用户名"
// @Param object body flow.StartTaskParams true "新增task参数"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task [post]
func (h *handler) StartTask(c *gin.Context) {
	user := c.GetHeader("StaffName")
	params := flow.NewStartTaskParams()
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	// 校验参数合法性
	//ok, msg, err := pkg.Domain.ValidateParam(user, params.FlowType, params.EchoFields)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
	//	return
	//}
	//if !ok {
	//	c.JSON(http.StatusBadRequest, frame.NewBadRequest(errors.New(msg)))
	//	return
	//}
	a, err := h.service.StartTask(user, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// DescribeTask 获取任务
// @Summary task详情
// @version 1.0.0
// @tags 流程管理
// @description  根据taskID查询task详情
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param task_id path string true "task ID,流程ID"
// @Param use_cache query string true "当use_cache=false时不经过缓存，默认都先从缓存查询"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task/{task_id} [get]
func (h *handler) DescribeTask(c *gin.Context) {
	user := c.GetHeader("StaffName")
	stepID := c.Param("task_id")
	useCacheParam := c.Query("use_cache")

	useCache := true
	if useCacheParam == "false" {
		useCache = false
	}
	a, err := h.service.DescribeTask(user, stepID, useCache)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// SaveFieldValue 保存字段
// @Summary 保存字段
// @version 1.0.0
// @tags 流程管理
// @description  根据步骤ID修改字段值
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body object true "参数：{field_id: field_value}"
// @Param task_id path string true "task ID,任务ID"
// @Param step_id path string true "step ID,步骤ID"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/tasks/{task_id}/steps/{step_id}/field [put]
func (h *handler) SaveFieldValue(c *gin.Context) {
	user := c.GetHeader("StaffName")
	stepID := c.Param("step_id")
	taskID := c.Param("task_id")
	params := new(interface{})
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	//判断操作人是否有权限，仅步骤owner或管理员有权限修改字段值
	ok, err := h.service.IsUserStepOwner(user, taskID, stepID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	if !ok {
		if !pkg.Role.HasRolePermission(role.FlowAdminName, user) && !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
			c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
			return
		}
	}
	a, err := h.service.SaveFieldValue(user, stepID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// SubmitStep 提交步骤
// @Summary 提交步骤
// @version 1.0.0
// @tags 流程管理
// @description  根据step ID提交步骤，通常在保存字段后操作
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param step_id path string true "step ID,步骤ID"
// @Param task_id path string true "task ID,任务ID"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task/{task_id}/steps/{step_id}/submit [put]
func (h *handler) SubmitStep(c *gin.Context) {
	user := c.GetHeader("StaffName")
	stepID := c.Param("step_id")
	taskID := c.Param("task_id")

	//判断操作人是否有权限，仅步骤owner或管理员有权限提交步骤
	ok, err := h.service.IsUserStepOwner(user, taskID, stepID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	if !ok {
		if !pkg.Role.HasRolePermission(role.FlowAdminName, user) && !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
			c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
			return
		}
	}
	a, err := h.service.SubmitStep(user, taskID, stepID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// UpdateStepOwner 修改步骤审批人
// @Summary 修改步骤审批人
// @version 1.0.0
// @tags 流程管理
// @description  根据step ID提交步骤，通常在保存字段后操作
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param step_id path string true "step ID,步骤ID"
// @Param task_id path string true "task ID,任务ID"
// @Param object body []string true "修改的审批人列表"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task/{task_id}/steps/{step_id}/update_step_owner [put]
func (h *handler) UpdateStepOwner(c *gin.Context) {
	user := c.GetHeader("StaffName")
	stepID := c.Param("step_id")
	taskID := c.Param("task_id")
	// 获取请求参数
	var owners []string
	err := c.BindJSON(&owners)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	t, err := pkg.Flow.DescribeTask(user, taskID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	//不能将审批人设置为流程发起人
	for _, name := range owners {
		if t.Creator == name {
			c.JSON(http.StatusBadRequest, frame.NewBadRequest(errors.New("不能将审批人设置为流程发起人")))
			return
		}
	}
	//判断操作人是否有权限，仅流程发起人或管理员有权限修改步骤审批人
	if user != t.Creator && !pkg.Role.HasRolePermission(role.FlowAdminName, user) &&
		!pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	//操作
	a, err := h.service.UpdateStepOwner(user, taskID, stepID, owners)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// StopTask 中止流程
// @Summary 中止流程
// @version 1.0.0
// @tags 流程管理
// @description  根据taskID中止步骤，只有creator和管理员可以操作
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param task_id path string true "step ID,步骤ID"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task/{task_id}/stop [put]
func (h *handler) StopTask(c *gin.Context) {
	user := c.GetHeader("StaffName")
	taskID := c.Param("task_id")
	params := new(interface{})
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	//判断操作人是否有权限，仅管理员有权限stop流程
	if !pkg.Role.HasRolePermission(role.FlowAdminName, user) && !pkg.Role.HasRolePermission(role.SuperAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	a, err := h.service.StopTask(user, taskID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// RevokeTask 撤回流程
// @Summary 撤回流程
// @version 1.0.0
// @tags 流程管理
// @description  根据taskID中止步骤，只有creator可以操作
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param task_id path string true "step ID,步骤ID"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/task/{task_id}/revoke [put]
func (h *handler) RevokeTask(c *gin.Context) {
	user := c.GetHeader("StaffName")
	taskID := c.Param("task_id")
	params := new(interface{})
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	//判断操作人是否有权限，仅流程发起人可以撤回
	ok, err := h.service.IsUserTaskCreator(user, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	a, err := h.service.StopTask(user, taskID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// MyApplication 我的申请
// @Summary 我的申请
// @version 1.0.0
// @tags 流程列表
// @description  查询我发起所有流程，支持各种参数
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body flow.QueryTaskParams true "查询task参数"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/tasks/my_application [post]
func (h *handler) MyApplication(c *gin.Context) {
	user := c.GetHeader("StaffName")
	params := flow.QueryTaskParams{}
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	params.Creator = user
	a, err := h.service.QueryTask(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// MyApprove 我的审批
// @Summary 我的审批
// @version 1.0.0
// @tags 流程列表
// @description  查询由我审批的流程，包括当前待我审批，以及我审批完成的流程
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body flow.QueryTaskParams true "查询task参数"
// @Param done query string true "当done=true查询已经审批完成的流程，否则查询当前待我审批的列表"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/tasks/my_approve [post]
func (h *handler) MyApprove(c *gin.Context) {
	user := c.GetHeader("StaffName")
	params := flow.QueryTaskParams{}
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	// 当附带query参数done=true的时候，查询已经完成审批，且step_owner为我的task_list
	status := c.Query("done")
	if status == "true" {
		params.StepOwner = user
	} else {
		params.DoingStepOwner = user
	}
	a, err := h.service.QueryTask(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// AllTaskList 流程列表
// @Summary 流程列表
// @version 1.0.0
// @tags 流程列表
// @description  查询所有流程列表，需要管理员权限
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param object body flow.QueryTaskParams true "查询task参数"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/tasks [post]
func (h *handler) AllTaskList(c *gin.Context) {
	user := c.GetHeader("StaffName")
	params := flow.QueryTaskParams{}
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	if !pkg.Role.HasRolePermission(role.SuperAdminName, user) && !pkg.Role.HasRolePermission(role.FlowAdminName, user) {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("permission deny")))
		return
	}
	a, err := h.service.QueryTask(user, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// AllFlowList 流程类型列表
// @Summary 流程类型列表
// @version 1.0.0
// @tags 流程列表
// @description  查询所有流程类型，为发起流程查询流程提供参数
// @Produce  json
// @Param StaffName header string true "用户名"
// @Success 200 {object} frame.AuthResponse
// @Router /flows [get]
func (h *handler) AllFlowList(c *gin.Context) {
	user := c.GetHeader("StaffName")
	a, err := h.service.GetFlowList(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewInternalError(err))
		return
	}
	c.JSON(http.StatusOK, frame.NewResponse(a))
}

// StatusNotify 状态推送
// @Summary 状态推送
// @version 1.0.0
// @tags 状态推送
// @description  QFlow状态推送
// @Produce  json
// @Param StaffName header string true "用户名"
// @Param _qflow_task_id query string true "推送ID"
// @Success 200 {object} frame.AuthResponse
// @Router /flow/status_notify [get]
func (h *handler) StatusNotify(c *gin.Context) {
	taskID := c.Query("_qflow_task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(errors.New("bad task id")))
		return
	}
	go h.service.NotifyTask(taskID)
	c.JSON(http.StatusOK, frame.NewResponse("query ok"))
}

// TaskAutoOperate 自动实施
func (h *handler) TaskAutoOperate(c *gin.Context) {
	qflowToken := c.GetHeader("Qflow-Token")
	if qflowToken != config.C().QFlow.AutoOperateToken {
		c.JSON(http.StatusForbidden, frame.NewPermissionDenied(errors.New("bad qflow token")))
		return
	}
	params := flow.OperateTaskParams{}
	err := c.BindJSON(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, frame.NewBadRequest(err))
		return
	}
	taskID := strconv.Itoa(params.TaskID)
	result, err := h.service.TaskAutoOperate(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, frame.NewFlowFailResponse(err.Error()))
	}
	resString, errors := json.Marshal(result)
	if errors != nil {
		c.JSON(http.StatusInternalServerError, frame.NewFlowFailResponse(errors.Error()))
	} else {
		c.JSON(http.StatusOK, frame.NewFlowResponse(string(resString)))
	}

}
