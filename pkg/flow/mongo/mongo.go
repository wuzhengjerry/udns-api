// Package mongo 审批流程
package mongo

import (
	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Service = &service{}
)

// service 服务集
type service struct {
	col *mongo.Collection
	l   *logrus.Logger
}

// Config 配置
func (s *service) Config() error {
	db := conf.C().Mongo.GetDB()
	col := db.Collection("task")

	s.col = col
	s.l = log.C()
	return nil
}

//func init() {
//	var _ flow.Service = Service
//	pkg.RegistryService("flow", Service)
//}
