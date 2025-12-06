package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/cache/memory"
	"github.com/wuzhengjerry/udns-api/cache/redis"
	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/frame"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	_ "github.com/wuzhengjerry/udns-api/pkg/all"
)

var (
	confType string
	confFile string
	confEtcd string
)

// startCmd represents the start command
var serviceCmd = &cobra.Command{
	Use:   "start",
	Short: "udns-web api服务",
	Long:  `udns-web api服务`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化全局变量
		if err := loadGlobalConfig(confType); err != nil {
			return err
		}

		// 初始化全局日志配置
		if err := log.SetGlobal(); err != nil {
			return err
		}
		log.C().Info("[START] load log success")

		// 加载缓存
		if err := loadCache(); err != nil {
			log.C().Error("[START] load cache failed:" + err.Error())
			return err
		}
		log.C().Info("[START] load cache success")

		// load mongo
		if err := conf.C().InitGloabl(); err != nil {
			log.C().Error("[START] get mongo conn failed:" + err.Error())
			return err
		}
		log.C().Info("[START] get mongo conn success")

		// 初始化服务层
		if err := pkg.InitService(); err != nil {
			log.C().Error("[START] init service failed:" + err.Error())
			return err
		}
		log.C().Info("[START] init service success")

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

		// 初始化服务
		svr, err := newService()
		if err != nil {
			log.C().Info("[START] new service failed")
			return err
		}

		// 等待信号处理
		go svr.waitSign(ch)

		// 启动服务
		if err := svr.start(); err != nil {
			if !strings.Contains(err.Error(), "http: Server closed") {
				log.C().Error("[START] start http server failed")
				return err
			}
		}

		return nil
	},
}

// newService 新建服务
func newService() (*service, error) {
	http := frame.NewHTTPService()

	svr := &service{
		http: http,
		log:  log.C(),
	}

	return svr, nil
}

// service 服务结构
type service struct {
	http *frame.HTTPService

	log  *logrus.Logger
	stop context.CancelFunc
}

// start 服务启动
func (s *service) start() error {
	//s.log.Infof("loaded services: %v", "nothing")
	return s.http.Start()
}

// loadGlobalConfig config 为全局变量, 只需要load 即可全局可用户
func loadGlobalConfig(configType string) error {
	// 配置加载
	switch configType {
	case "file":
		err := conf.LoadConfigFromToml(confFile)
		if err != nil {
			return err
		}
	case "env":
		err := conf.LoadConfigFromEnv()
		if err != nil {
			return err
		}
		return nil
	case "etcd":
		return errors.New("not implemented")
	default:
		return errors.New("unknown config type")
	}

	return nil
}

// loadCache 加载缓存
func loadCache() error {
	c := conf.C()
	// 设置全局缓存
	switch c.Cache.Type {
	case "memory", "":
		ins := memory.NewCache(c.Cache.Memory)
		cache.SetGlobal(ins)
		log.C().Info("[start] use cache in local memory")
	case "redis":
		ins := redis.NewCache(c.Cache.Redis)
		cache.SetGlobal(ins)
		log.C().Info("[start] use redis to cache")
	default:
		return fmt.Errorf("unknown cache type: %s", c.Cache.Type)
	}

	return nil
}

// waitSign 等待信号
func (s *service) waitSign(sign chan os.Signal) {
	for {
		select {
		case sg := <-sign:
			switch v := sg.(type) {
			default:
				s.log.Infof("receive signal '%v', start graceful shutdown", v.String())
				if err := s.http.Stop(); err != nil {
					s.log.Errorf("[STOP] graceful shutdown err: %s, force exit", err)
				}
				s.log.Infof("[STOP] module stop complete")
				return
			}
		}
	}
}

// init 初始化
func init() {
	serviceCmd.Flags().StringVarP(&confType, "config-type", "t", "file", "the module config type [file/env/etcd]")
	serviceCmd.Flags().StringVarP(&confFile, "config-file", "f", "etc/udns-api.toml", "the module config from file")
	RootCmd.AddCommand(serviceCmd)
}
