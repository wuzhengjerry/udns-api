package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	_ "github.com/wuzhengjerry/udns-api/pkg/all"
)

// startCmd represents the start command
var dailyNotifyCmd = &cobra.Command{
	Use:   "daily_notify",
	Short: "流程待办每日提醒",
	Long:  `流程待办每日提醒`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化全局变量
		if err := loadGlobalConfig(confType); err != nil {
			return err
		}

		// 初始化全局日志配置
		if err := log.SetGlobal(); err != nil {
			return err
		}
		log.C().Info("[NOTIFY] load log success")

		// 加载缓存
		if err := loadCache(); err != nil {
			log.C().Error("[NOTIFY] load cache failed:" + err.Error())
			return err
		}
		log.C().Info("[NOTIFY] load cache success")

		// load mongo
		if err := conf.C().InitGloabl(); err != nil {
			log.C().Error("[NOTIFY] get mongo conn failed:" + err.Error())
			return err
		}
		log.C().Info("[NOTIFY] get mongo conn success")

		// 初始化服务层
		if err := pkg.InitService(); err != nil {
			log.C().Error("[NOTIFY] init service failed:" + err.Error())
			return err
		}
		log.C().Info("[NOTIFY] init service success")

		// 初始化管理员信息
		if err := pkg.Flow.DoingTaskDailyNotify(); err != nil {
			log.C().Error("[NOTIFY] DoingTaskDailyNotify failed:" + err.Error())
			return err
		}
		log.C().Info("[NOTIFY] doing task daily notify success")

		return nil
	},
}

// init 初始化
func init() {
	dailyNotifyCmd.Flags().StringVarP(&confType, "config-type", "t", "file", "the module config type [file/env/etcd]")
	dailyNotifyCmd.Flags().StringVarP(&confFile, "config-file", "f", "etc/udns-api.toml", "the module config from file")
	RootCmd.AddCommand(dailyNotifyCmd)
}
