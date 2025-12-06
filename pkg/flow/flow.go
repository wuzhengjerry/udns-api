// Package flow 审批流操作结构体
package flow

// Flow 流程
type Flow struct {
	ID                int32  `json:"id"`
	Affiche           string `json:"affiche"`
	Name              string `json:"name"`
	PersonalNotifyWay string `json:"personal_notify_way"`
}

// Task 流程任务
type Task struct {
	ID              int32       `json:"id" bson:"id"`                               // ID
	Creator         string      `json:"creator" bson:"creator"`                     // 流程创建人
	DeploymentID    int32       `json:"deployment_id" bson:"deployment_id"`         //deployment ID
	Description     string      `json:"description" bson:"description"`             // 描述
	FinishTime      string      `json:"finish_time" bson:"finish_time"`             //结束时间
	Flow            *Flow       `json:"flow" bson:"flow"`                           // 对应流程
	FlowID          int32       `json:"flow_id" bson:"flow_id"`                     // 流程ID
	IntervalTip     string      `json:"interval_tip" bson:"interval_tip"`           //
	Memo            interface{} `json:"memo" bson:"memo"`                           //
	Name            string      `json:"name" bson:"name"`                           // 流程名
	NotifySessionID string      `json:"notify_session_id" bson:"notify_session_id"` //
	Permission      string      `json:"permission" bson:"permission"`               //权限
	PersonalTip     string      `json:"personal_tip" bson:"personal_tip"`
	PopStepList     *[]Step     `json:"pop_step_list" bson:"pop_step_list"`
	Progress        string      `json:"progress" bson:"progress"` // 进度
	RtxGroup        string      `json:"rtx_group" bson:"rtx_group"`
	StartTime       string      `json:"start_time" bson:"start_time"`       // 开始时间
	Status          string      `json:"status" bson:"status"`               // 状态
	Steps           []Step      `json:"steps" bson:"steps"`                 // 步骤
	SubFlowMsg      interface{} `json:"sub_flow_msg" bson:"sub_flow_msg"`   //
	TimeCost        string      `json:"time_cost" bson:"time_cost"`         // 耗时
	UnShowSteps     *[]Step     `json:"un_show_steps" bson:"un_show_steps"` //
	ViewOwner       string      `json:"view_owner" bson:"view_owner"`       // 可见用户
	//Xml             string      `json:"xml" bson:"xml"`                     //
}

// Step 步骤
type Step struct {
	ID            int32       `json:"id" bson:"id"`                           // ID
	Creator       string      `json:"creator" bson:"creator"`                 // 创建人
	Description   string      `json:"description" bson:"description"`         // 描述
	Fields        *[]Field    `json:"fields" bson:"fields"`                   // 字段
	FinishTime    string      `json:"finish_time" bson:"finish_time"`         // 结束时间
	GroupName     string      `json:"group_name" bson:"group_name"`           //
	Hold          bool        `json:"hold" bson:"hold"`                       //
	Link          string      `json:"link" bson:"link"`                       //
	Name          string      `json:"name" bson:"name"`                       // 步骤名称
	Owner         string      `json:"owner" bson:"owner"`                     // 所有人
	ShapeID       string      `json:"shape_id" bson:"shape_id"`               //
	StartTime     string      `json:"start_time" bson:"start_time"`           // 开始时间
	Status        string      `json:"status" bson:"status"`                   // 状态
	SubmitVarName string      `json:"submit_var_name" bson:"submit_var_name"` //
	TimeCost      string      `json:"time_cost" bson:"time_cost"`             // 耗时
	ToolLog       interface{} `json:"tool_log" bson:"tool_log"`               //
}

// Field 字段
type Field struct {
	ID                int32       `json:"id"`
	Description       string      `json:"description"`
	FieldNum          int32       `json:"field_num"`
	FieldType         string      `json:"field_type"`
	InputValue        string      `json:"input_value"`
	IsEdit            bool        `json:"is_edit"`
	IsEncrypt         string      `json:"is_encrypt"`
	IsOnTaskTop       string      `json:"is_on_task_top"`
	IsPushed          string      `json:"is_pushed"`
	IsRequired        string      `json:"is_required"`
	IsTaskName        string      `json:"is_task_name"`
	Link              string      `json:"link"`
	Name              string      `json:"name"`
	Note              string      `json:"note"`
	Option            string      `json:"option"`
	OptionDescription string      `json:"option_description"`
	OptionJson        interface{} `json:"option_json"`
	SubmitVarName     string      `json:"submit_var_name"`
}

// StartTaskParams 发起 qflow 请求参数
type StartTaskParams struct {
	FlowID         int32             `json:"flow_id"`
	FlowType       string            `json:"flow_type"`
	Basic          *Basic            `json:"basic"`
	EchoFields     map[string]string `json:"echo_fields"`
	EchoStepOwners map[string]string `json:"echo_step_owners"`
}

// NewStartTaskParams 拼接 qflow 请求参数
func NewStartTaskParams() *StartTaskParams {
	return &StartTaskParams{
		Basic:          &Basic{},
		EchoFields:     make(map[string]string),
		EchoStepOwners: make(map[string]string),
	}
}

// Basic 任务名称
type Basic struct {
	TaskName  string `json:"task_name"`
	ChildItem string `json:"child_item"`
}

// QueryTaskParams 查询 qflow 参数
type QueryTaskParams struct {
	Page            int32    `json:"page"`
	PageSize        int32    `json:"page_size"`
	CreateTimeEnd   string   `json:"create_time_end"`
	CreateTimeStart string   `json:"create_time_start"`
	Creator         string   `json:"creator"`
	DoingStepOwner  string   `json:"doing_step_owner"`
	FieldInputValue string   `json:"field_input_value"`
	FinishTimeEnd   string   `json:"finish_time_end"`
	FinishTimeStart string   `json:"finish_time_start"`
	FlowID          []string `json:"flow_id"`
	Name            string   `json:"name"`
	Status          []string `json:"status"`
	StepOwner       string   `json:"step_owner"`
}

// QueryTaskResp 查询 qflow 任务返回结果
type QueryTaskResp struct {
	Count      int32  `json:"count"`
	Results    []Task `json:"results"`
	Next       string `json:"next"`
	Permission bool   `json:"permission"`
	Previous   string `json:"previous"`
}

// QueryFlowResp 查询 qflow 返回结果
type QueryFlowResp struct {
	Count      int32  `json:"count"`
	Results    []Flow `json:"results"`
	Next       string `json:"next"`
	Permission bool   `json:"permission"`
	Previous   string `json:"previous"`
}

// OperateTaskParams 操作 qflow 任务
type OperateTaskParams struct {
	TaskID int `json:"task_id"`
}
