// Package openapi 后端调用操作结构体
package openapi

import "sync"

// QueryParam 查询参数
type QueryParam struct {
	User        string   `json:"user"`        //OpenApi 用户
	Domains     []string `json:"domains"`     // 有权限的域名
	Name        []string `json:"name"`        // 域名名称
	GroupID     []int    `json:"group_id"`    // 所属组织组ID
	DeptID      []int    `json:"dept_id"`     // 所属组织部门ID
	Owner       []string `json:"owner"`       // 域名所有人
	Operator    []string `json:"operator"`    // 域名权限人
	Description string   `json:"description"` // 描述信息
	Regulator   []string `json:"regulator"`   // 域名管理者：为责任人或权限人
	Limit       int64    `json:"limit"`       // 返回数量
	Offset      int64    `json:"offset"`      // 偏移量
}

// DomainConfig 域名配置
type DomainConfig struct {
	Content string `json:"content"` //解析内容
	Route   string `json:"route"`   //解析路线
	UType   int32  `json:"uType"`   //域名类型1，5，28
}

// DomainInfo 域名信息
type DomainInfo struct {
	Name         string          `json:"name"`          // 域名名称
	GroupID      int32           `json:"group_id"`      // 所属组织ID
	GroupName    string          `json:"group_name"`    // 所属组织名称
	DeptID       int32           `json:"dept_id"`       // 所属组织部门ID
	DeptName     string          `json:"dept_name"`     // 所属组织部门名称
	Owner        string          `json:"owner"`         // 域名所有人逗号隔开
	Operator     string          `json:"operator"`      // 域名所有人
	Description  string          `json:"description"`   // 描述信息
	CreatedTime  int64           `json:"created_time"`  // 创建时间
	UpdatedTime  int64           `json:"updated_time"`  // 更新时间
	UTTL         int32           `json:"uTTL"`          //uttl:300,600,1200,3600
	Config       []*DomainConfig `json:"config"`        //
	BusinessName string          `json:"business_name"` // 业务用途
}

// DelDomainParam 域名删除参数
type DelDomainParam struct {
	Name   []string `json:"name"`   //删除域名列表
	Reason string   `json:"reason"` //删除原因
}

// WaitGroup 异步结构体
type WaitGroup struct {
	workChan chan int
	swg      sync.WaitGroup
}

// NewPool 生成一个工作池, coreNum 限制
func NewPool(coreNum int) *WaitGroup {
	ch := make(chan int, coreNum)
	return &WaitGroup{
		workChan: ch,
		swg:      sync.WaitGroup{},
	}
}

// Add 添加
func (ap *WaitGroup) Add(num int) {
	for i := 0; i < num; i++ {
		ap.workChan <- i
		ap.swg.Add(1)
	}
}

// Done 完结
func (ap *WaitGroup) Done() {
LOOP:
	for {
		select {
		case <-ap.workChan:
			break LOOP
		}
	}
	ap.swg.Done()
}

// Wait 等待
func (ap *WaitGroup) Wait() {
	ap.swg.Wait()
}
