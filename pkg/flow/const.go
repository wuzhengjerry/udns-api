package flow

import "time"

const (
	OneDayTTL = 24 * time.Hour

	BasePermGroup = "base_perm_group"

	SystemUserName = "system"

	TaskCacheKeyPrefix = "task_"

	FlowListCacheKey = "qflow_flow_list"
	// Task Status
	DoneStatus  = "done"
	TodoStatus  = "todo"
	DoingStatus = "doing"
	StopStatus  = "stop"
	PauseStatus = "pause"
	// Flow name 区分不同类型流程
	ApplyDomainsFlow      = "域名申请"
	ApplyDomainsFlowCSIG      = "域名申请(CSIG)"
	DeleteDomainsFlow     = "域名删除"
	ApplyOwnerPermFlow    = "域名责任人申请"
	ApplyOperatorPermFlow = "域名配置权限申请"
	TransferDomainsFlow   = "转出域名责任人"
	// 直属leader审批步骤名
	LeaderApproveStepName = "直属leader审批"
)
