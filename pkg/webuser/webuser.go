// Package webuser 用户管理
package webuser

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebUser 用户
type WebUser struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	Name        string             `json:"name" bson:"name"`                 // 调用用户
	Description string             `json:"description" bson:"description"`   // 调用描述
	Domain      []string           `json:"domain" bson:"domain"`             // 管理域名
	Operator    string             `json:"operator" bson:"operator"`         // 申请人
	Token       string             `json:"token" bson:"token"`               // 鉴权token
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 创建时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// Validate 用户验证
func (z *WebUser) Validate() error {
	if z.Name == "" {
		return errors.New("name is required")
	}
	if z.Operator == "" {
		return errors.New("operator is required")
	}
	if len(z.Domain) == 0 {
		return errors.New("domain is required")
	}
	return nil
}
