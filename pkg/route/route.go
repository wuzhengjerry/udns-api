// Package route 路由结构体
package route

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Route 线路
type Route struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	Name        string             `json:"name" bson:"name"`                 // 线路名称
	RouteID     string             `json:"route_id" bson:"route_id"`         // 线路ID
	Operator    string             `json:"operator" bson:"operator"`         // 负责人
	Enabled     bool               `json:"enabled" bson:"enabled"`           // 是否启用
	Description string             `json:"description" bson:"description"`   // 描述信息
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 创建时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// Validate 验证
func (r *Route) Validate() error {
	if r.Name == "" {
		return errors.New("route name is required")
	}
	if r.RouteID == "" {
		return errors.New("route id is required")
	}
	if r.Operator == "" {
		return errors.New("route operator is required")
	}
	return nil
}
