// Package mongo 路由操作模块
package mongo

import (
	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/route"
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
	col := db.Collection("route")

	s.col = col
	s.l = log.C()
	return nil
}

// init 初始化
func init() {
	var _ route.Service = Service
	pkg.RegistryService("route", Service)
}
