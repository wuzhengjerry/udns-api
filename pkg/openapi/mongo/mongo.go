// Package mongo 后端调用
package mongo

import (
	"github.com/sirupsen/logrus"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	"github.com/wuzhengjerry/udns-api/pkg/openapi"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Service = &service{}
)

// service 组件集
type service struct {
	col  *mongo.Collection
	uu   *mongo.Collection
	logs *mongo.Collection
	l    *logrus.Logger
	//d   *domain.Service
}

// Config 配置
func (s *service) Config() error {
	db := conf.C().Mongo.GetDB()
	col := db.Collection("domain")
	uu := db.Collection("webuser")
	logs := db.Collection("syncapilog")
	s.col = col
	s.uu = uu
	s.logs = logs
	s.l = log.C()
	//s.d = pkg.domain
	return nil
}

// init 初始化
func init() {
	var _ openapi.Service = Service
	pkg.RegistryService("openapi", Service)
}
