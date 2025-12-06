package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wuzhengjerry/udns-api/conf"
	log "github.com/wuzhengjerry/udns-api/logs"
	"github.com/wuzhengjerry/udns-api/pkg"
	_ "github.com/wuzhengjerry/udns-api/pkg/all"
	"github.com/wuzhengjerry/udns-api/pkg/role"
)

const (
	InitUser = "jerryzwu;charleyluo;lioncoldlin"
)

// initCmd represents the start command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化任务",
	Long:  `初始化基础角色，权限等信息`,
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

		// 初始化管理员信息
		if err := InitAdmin(); err != nil {
			log.C().Error("[START] init admin failed:" + err.Error())
			return err
		}
		log.C().Info("[START] init admin success")

		return nil
	},
}

// InitAdmin 初始化
func InitAdmin() error {
	superAdmin := role.Role{
		Name:        role.SuperAdminName,
		Users:       InitUser,
		Description: "全局管理员，拥有所有权限",
	}
	err := pkg.Role.InsertRole(&superAdmin)
	if err != nil {
		return err
	}

	flowAdmin := role.Role{
		Name:        role.FlowAdminName,
		Users:       InitUser,
		Description: "流程管理员，拥有流程审核和流程管理权限",
	}
	err = pkg.Role.InsertRole(&flowAdmin)
	if err != nil {
		return err
	}

	return nil
}

// init 初始化
func init() {
	initCmd.Flags().StringVarP(&confType, "config-type", "t", "file", "the module config type [file/env/etcd]")
	initCmd.Flags().StringVarP(&confFile, "config-file", "f", "etc/udns-api.toml", "the module config from file")
	RootCmd.AddCommand(initCmd)
}
