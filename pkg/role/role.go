// Package role 角色结构体
package role

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role 角色
type Role struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	Name        RoleName           `json:"name" bson:"name"`                 // 线路名称
	Users       string             `json:"users" bson:"users"`               // 关联用户，分号相隔
	Description string             `json:"description" bson:"description"`   // 描述信息
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 创建时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// RoleName 角色名称
type RoleName string

// Validate 校验
func (r *Role) Validate() error {
	if r.Name == "" {
		return errors.New("role name is required")
	}
	if r.Name != SuperAdminName && r.Name != FlowAdminName {
		return errors.New("bad role name")
	}
	return nil
}
