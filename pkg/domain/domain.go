// Package domain 域名操作结构体
package domain

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Domain 域名信息
type Domain struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`                      // ID
	Name         string             `json:"name" bson:"name"`                   // 域名名称
	BusinessID   string             `json:"business_id" bson:"business_id"`     // CMDB业务ID
	BusinessName string             `json:"business_name" bson:"business_name"` // CMDB业务名
	DeptID       int                `json:"dept_id" bson:"dept_id"`             // 所属组织ID
	DeptName     string             `json:"dept_name" bson:"dept_name"`         // 所属组织名称
	GroupID      int                `json:"group_id" bson:"group_id"`           // 所属组织ID
	GroupName    string             `json:"group_name" bson:"group_name"`       // 所属组织名称
	Owner        string             `json:"owner" bson:"owner"`                 // 域名所有人逗号隔开
	Operator     string             `json:"operator" bson:"operator"`           // 域名所有人
	Description  string             `json:"description" bson:"description"`     // 描述信息
	ITFlag       string             `json:"it_flag" bson:"it_flag"`             // 标识是否同步IT数据
	RouteConfigs []*RouteConfig     `json:"route_configs" bson:"-"`             // 线路配置
	Deleted      bool               `json:"deleted" bson:"deleted"`             // 是否删除
	CreatedTime  int64              `json:"created_time" bson:"created_time"`   // 创建时间
	UpdatedTime  int64              `json:"updated_time" bson:"updated_time"`   // 恢复时间
	DeletedTime  int64              `json:"deleted_time" bson:"deleted_time"`   // 删除时间
}

// RouteConfig 线路配置
type RouteConfig struct {
	Name         string `json:"name"`                               // 域名
	Operator     string `json:"operator"`                           //责任人
	Route        string `json:"route"`                              // 线路ID
	UType        int32  `json:"uType"`                              // A,AAAA,CNAME等几类
	UTTL         int32  `json:"uTTL"`                               // TTL
	SMasterList  string `json:"sMasterList"`                        // 是否启用
	Description  string `json:"description"`                        // 描述信息
	BusinessId   string `json:"business_id"`                        //业务ID
	Zone         string `json:"zone"`                               //二级域
	ZoneId       int32  `json:"zone_id"`                            //zoneID
	BusinessName string `json:"business_name" bson:"business_name"` // 业务用途
}

// DomainLog 域名操作日志
type DomainLog struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	DomainName  string             `json:"domain_name" bson:"domain_name"`   // 域名ID
	Method      string             `json:"method" bson:"method"`             // 方法
	Message     string             `json:"message" bson:"message"`           // 信息
	Operator    string             `json:"operator" bson:"operator"`         // 操作人
	Version     int64              `json:"version" bson:"version"`           // 版本
	Description string             `json:"description" bson:"description"`   // 描述信息
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 创建时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// SyncApiLog 同步日志
type SyncApiLog struct { //调用it\udns_web\udns_ori接口日志
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	ApiType     string             `json:"api_type" bson:"api_type"`         // 接口类型:udns;udns_web,it
	DomainName  string             `json:"domain_name" bson:"domain_name"`   // 域名ID
	Method      string             `json:"method" bson:"method"`             // 方法
	Message     string             `json:"message" bson:"message"`           // 调用参数
	Success     string             `json:"success" bson:"success"`           // 调用成功，是或否，没有用其他数据类型为了方便查询
	Msg         string             `json:"msg" bson:"msg"`                   //返回信息
	Operator    string             `json:"operator" bson:"operator"`         // 操作人ßß
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 创建时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// IdcData idc数据
type IdcData struct {
	SIPList  string `json:"sIPList"`
	UIdcFlag int32  `json:"uIdcFlag"`
	UIdcId   int32  `json:"uIdcId"`
}

// RRData rr记录
type RRData struct {
	SLocalDns    string `json:"sLocalDns"`
	SMasterIDCID string `json:"sMasterIDCID"`
	SMasterList  string `json:"sMasterList"`
	SSlaveIDCID  string `json:"sSlaveIDCID"`
	SSlaveList   string `json:"sSlaveList"`
	UCountry     int32  `json:"uCountry"`
	UFlag        int32  `json:"uFlag"`
	UISP         int32  `json:"uISP"`
	UProvince    int32  `json:"uProvince"`
	UTTL         int32  `json:"uTTL"`
	UType        int32  `json:"uType"`
}

// QueryDomainParam 域名查询参数
type QueryDomainParam struct {
	NSysId  int64  `json:"nSysId"`
	SArea   string `json:"sArea"`
	SDomain string `json:"sDomain"`
}

// BatchQueryDomainParam 域名批量查询参数
type BatchQueryDomainParam struct {
	NSysId  int64  `json:"nSysId"`
	SDomain string `json:"sDomain"`
}

// DelDomainParam 域名删除参数
type DelDomainParam struct {
	NSysId    int64  `json:"nSysId"`
	SArea     string `json:"sArea"`
	SDomain   string `json:"sDomain"`
	SProposer string `json:"sProposer"`
}

// ModDomainParam 域名修改参数
type ModDomainParam struct {
	Idc         []IdcData `json:"Idc"`
	RR          []RRData  `json:"RR"`
	SComment    string    `json:"sComment"`
	SCountry    string    `json:"sCountry"`
	SDomainName string    `json:"sDomainName"`
	SISP        string    `json:"sISP"`
	SZoneName   string    `json:"sZoneName"`
	UIDCCount   int32     `json:"uIDCCount"`
	URRCount    int32     `json:"uRRCount"`
	UTTL        int32     `json:"uTTL"`
	UZoneID     int32     `json:"uZoneID"`
	NSysId      int64     `json:"nSysId"`
	SProposer   string    `json:"sProposer"`
	SOwner      string    `json:"sOwner"`
	SArea       string    `json:"sArea"`
}

// ITDomainInfo IT域名信息
type ITDomainInfo struct { //IT接口所需域名参数
	AutoRedirectOption    string `json:"auto_redirect_option"`     //域名跳转至新域名
	Domain                string `json:"domain"`                   //域名
	OldAutoRedirectOption string `json:"old_auto_redirect_option"` //旧域名跳转至新域名
	OldPServer            string `json:"old_p_server"`             //旧域名解析IP或者域名（A,CNAME,AAAA）
	OldPaasSyncOption     bool   `json:"old_paas_sync_option"`     //旧同步原域名PaaS权限,默认false
	OldTTL                int32  `json:"old_ttl"`                  //变更和删除必填
	OldVUserSyncOption    bool   `json:"old_v_user_sync_option"`   //旧外包访问权限平移至新域名, 默认false
	OperateType           string `json:"operatetype"`              //申请、变更、删除
	PServer               string `json:"p_server"`                 //域名解析IP或者域名（A,CNAME,AAAA）
	PaasSyncOption        string `json:"paas_sync_option"`
	Principal             string `json:"principal"`
	TTL                   int32  `json:"ttl"`
	VUserSyncOption       string `json:"v_user_sync_option"` //外包访问权限平移至新域名
	YDomain               string `json:"y_domain"`           //原域名
	Source                int32  `json:"source"`             //来源,默认2

}

// ITParam IT参数
type ITParam struct {
	Applicant     string         `json:"Applicant"`   //员工ID，英文名即可
	ITOperator    string         `json:"ITOperator"`  //姓名，英文名即可
	CaseId        string         `json:"caseid"`      //流程ID
	Department    string         `json:"department"`  //申请人部门
	DomainInfo    []ITDomainInfo `json:"domain_info"` //域名实体
	ITAuditAdmin  string         `json:"it_audit_admin"`
	K2InstID      string         `json:"k2InstID"`       //k2工作流id
	K2sn          string         `json:"k2sn"`           //k2工作流编码
	OperationTime string         `json:"operation_time"` //操作时间
}

// ITApiResp IP请求响应
type ITApiResp struct {
	Data       string `json:"Data"`
	Msg        string `json:"Msg"`
	ExecStatus int32  `json:"ExecStatus"`
}

// ITRApiResp ITR请求响应
type ITRApiResp struct {
	ResolveAddr string `json:"resolve_addr"`
	DomainName  string `json:"domain_name"`
	TTL         int32  `json:"ttl"`
	Admin       string `json:"admin"`
}

// Message 返回消息
type Message struct {
	Success bool
	Message string
}

// OudnsResp UDN响应
type OudnsResp struct {
	ErrMsg string `json:"ErrMsg"`
	Result int32  `json:"Result"`
}

// Validate 域名验证
func (d *Domain) Validate() error {
	if d.Name == "" {
		return errors.New("domain name is required")
	}
	//if d.Owner == "" {
	//	return errors.New("domain owner is required")
	//}
	//if d.Operator == "" {
	//	return errors.New("domain operator is required")
	//}
	return nil
}

// LogValidate 域名日志验证
func (d *DomainLog) LogValidate() error {
	if d.DomainName == "" {
		return errors.New("domain name is required")
	}
	if d.Method == "" {
		return errors.New("domain name is required")
	}
	if d.Message == "" {
		return errors.New("domain name is required")
	}
	if d.Operator == "" {
		return errors.New("domain operator is required")
	}
	return nil
}
