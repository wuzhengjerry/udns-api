package flow

import "github.com/wuzhengjerry/udns-api/pkg/domain"

// Service 流程服务
type Service interface {
	StartTask(username string, params *StartTaskParams) (res int32, err error)
	DescribeTask(username, taskID string, useCache bool) (task *Task, err error)
	GetFlowList(username string) (res *QueryFlowResp, err error)
	SaveFieldValue(username, stepID string, params interface{}) (res interface{}, err error)
	SubmitStep(username, taskID, stepID string) (res interface{}, err error)
	UpdateStepOwner(username, taskID, stepID string, owners []string) (res interface{}, err error)
	StopTask(username, taskID string, params interface{}) (res interface{}, err error)
	QueryTask(username string, qt *QueryTaskParams) (res *QueryTaskResp, err error)
	IsUserTaskCreator(username, taskID string) (ok bool, err error)
	IsUserStepOwner(username, taskID, stepID string) (ok bool, err error)
	//QueryFlow(username string) (res interface{}, err error)
	NotifyTask(taskID string)
	DoingTaskDailyNotify() error
	TaskAutoOperate(taskID string) (result map[string]domain.Message, err error)
}
