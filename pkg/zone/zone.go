// Package zone 地域结构体
package zone

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Zone 二级zone
type Zone struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`                    // ID
	Name        string             `json:"name" bson:"name"`                 // 二级zone域名
	ZoneID      int                `json:"zone_id" bson:"zone_id"`           // 二级zone id
	Operator    string             `json:"operator" bson:"operator"`         // 负责人
	Enabled     bool               `json:"enabled" bson:"enabled"`           // 是否启用
	Description string             `json:"description" bson:"description"`   // 描述信息
	Deleted     bool               `json:"deleted" bson:"deleted"`           // 是否删除
	CreatedTime int64              `json:"created_time" bson:"created_time"` // 触发时间
	UpdatedTime int64              `json:"updated_time" bson:"updated_time"` // 恢复时间
	DeletedTime int64              `json:"deleted_time" bson:"deleted_time"` // 删除时间
}

// Validate zone验证
func (z *Zone) Validate() error {
	if z.Name == "" {
		return errors.New("zone name is required")
	}
	//if len(strings.Split(z.Name, ".")) != 3 {
	//	return errors.New("bad zone name")
	//}
	//if !strings.HasSuffix(z.Name, ".com.") {
	//	return errors.New("bad zone name has bad suffix")
	//}
	return nil
}
