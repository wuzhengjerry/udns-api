// Package mongo 域名操作
package mongo

import (
	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Service = &service{}
)

// service 服务集
type service struct {
	col    *mongo.Collection
	logs   *mongo.Collection
	itlogs *mongo.Collection
	l      *logrus.Logger
	//d   *domain.Service
}

// Config 配置
func (s *service) Config() error {
	db := conf.C().Mongo.GetDB()
	col := db.Collection("domain")
	logs := db.Collection("domainlog")
	itLogs := db.Collection("syncapilog")
	s.col = col
	s.logs = logs
	s.itlogs = itLogs
	s.l = log.C()
	//s.d = pkg.domain
	return nil
}

// init 初始化
func init() {
	var _ domain.Service = Service
	pkg.RegistryService("domain", Service)
}
