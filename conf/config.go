// Package conf 配置模块
package conf

import (
	"context"
	"fmt"
	"time"

	"github.com/wuzhengjerry/udns-api/cache/memory"
	"github.com/wuzhengjerry/udns-api/cache/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoDefaultSource = "admin"
)

var (
	mgoclient *mongo.Client
)

// newConfig 默认配置
func newConfig() *Config {
	return &Config{
		App:       newDefaultAPP(),
		Log:       newDefaultLog(),
		Mongo:     newDefaultMongoDB(),
		Cache:     newDefaultCache(),
		QFlow:     newDefaultQFlow(),
		BQueue:    newDefaultBqueue(),
		Nops:      newDefaultNops(),
		FlyPigeon: newDefaultFlyPigeon(),
		TOF:       newDefaultTOF(),
		Email:     newDefaultEmail(),
		OUDNS:     newDefaultUNDS(),  // 原生UNDS接口 OriginalUDNS
		ITApi:     newDefaultITApi(), //原生IT接口
	}
}

// Config 应用配置
type Config struct {
	App       *app       `toml:"app"`
	Log       *log       `toml:"log"`
	Mongo     *mongodb   `toml:"mongo"`
	Cache     *_cache    `toml:"cache"`
	QFlow     *qflow     `toml:"qflow"`
	BQueue    *bqueue    `toml:"bqueue"`
	Nops      *nops      `toml:"nops"`
	FlyPigeon *flyPigeon `toml:"nops"`
	TOF       *tof       `toml:"tof"`
	Email     *email     `toml:"email"`
	OUDNS     *oudns     `toml:"oudns"`
	ITApi     *itapi     `toml:"itapi"`
}

// InitGloabl 注入全局变量
func (c *Config) InitGloabl() error {
	// 加载全局配置单例
	global = c

	// 加载全局数据量单例
	mclient, err := c.Mongo.getClient()
	if err != nil {
		return err
	}
	mgoclient = mclient
	return nil
}

// email 邮件配置
type email struct {
	ServerHost string `toml:"server_host" env:"UDNS_EMAIL_SERVER_HOST"`
	ServerPort int    `toml:"server_port" env:"UDNS_EMAIL_SERVER_PORT"`
	FromEmail  string `toml:"from_email" env:"UDNS_EMAIL_FROM_EMAIL"`
	Passwd     string `toml:"passwd" env:"UDNS_EMAIL_PASSWD"`
}

// newDefaultEmail 默认邮件配置
func newDefaultEmail() *email {
	return &email{
		ServerHost: "smtp@tencent.com",
		ServerPort: 465,
		FromEmail:  "udns@tencent.com",
		Passwd:     "",
	}
}

// bqueue 队列配置
type bqueue struct {
	Addr     string `toml:"addr" env:"UDNS_BQUEUE_ADDR"`
	Topic    string `toml:"topic" env:"UDNS_BQUEUE_TOPIC"`
	Group    string `toml:"group" env:"UDNS_BQUEUE_GROUP"`
	Version  string `toml:"version" env:"UDNS_BQUEUE_VERSION"`
	Assignor string `toml:"assignor" env:"UDNS_BQUEUE_ASSIGNOR"`
	Oldest   bool   `toml:"oldest" env:"UDNS_BQUEUE_OLDEST"`
}

// newDefaultBqueue 默认队列配置
func newDefaultBqueue() *bqueue {
	return &bqueue{
		Addr:     "",
		Topic:    "UDNS_center_test",
		Group:    "UDNS",
		Version:  "2.4.2",
		Assignor: "roundrobin",
		Oldest:   true,
	}
}

// nops 智能网关
type nops struct {
	WebapiHost string `toml:"webapi_host" env:"UDNS_NOPS_WEBAPI_HOST"`
	IGateToken string `toml:"i_gate_token" env:"UDNS_IGATE_TOKEN"`
}

// newDefaultNops 默认智能网关
func newDefaultNops() *nops {
	return &nops{
		WebapiHost: "http://test.nops.woa.com",
		IGateToken: "Fwn4ylkicdewMp9z1swLzYSbGu9FhoVT",
	}
}

// qflow 审批流
type qflow struct {
	Addr             string `toml:"addr" env:"UDNS_QFLOW_ADDR"`
	Token            string `toml:"token" env:"UDNS_QFLOW_TOKEN"`
	ProjectID        int32  `toml:"project_id" env:"UDNS_QFLOW_PROJECT_ID"`
	ProjectName      string `toml:"project_name" env:"UDNS_QFLOW_PROJECT_NAME"`
	AutoOperateToken string `toml:"auto_operate_token" env:"AUTO_OPERATE_TOKEN"`
}

// newDefaultQFlow 默认审批流配置
func newDefaultQFlow() *qflow {
	return &qflow{
		Addr:             "http://dev.qflow.woa.com",
		Token:            "4bbbc3c0-9b2d-11eb-b9f5-ba3fe58b92c7",
		ProjectID:        148,
		ProjectName:      "udns-web",
		AutoOperateToken: "udns-dev",
	}
}

// oudns udns配置
type oudns struct {
	Addr          string `toml:"addr" env:"UDNS_ADDR"`
	SystemID      int64  `toml:"system_id" env:"SYSTEM_ID"`
	Proposer      string
	AddApi        string
	AddZoneApi    string
	ModApi        string
	DelApi        string
	QueryApi      string
	BatchQueryApi string
	BatchModApi   string
}

// itapi 配置
type itapi struct { //original_it_api
	Addr   string `toml:"addr" env:"IT_ADDR"`
	RAddr  string `toml:"addr" env:"IT_RADDR"`
	ModApi string
	RApi   string
	Token  string `env:"IT_TOKEN"`
}

// newDefaultITApi 默认ITApi配置
func newDefaultITApi() *itapi {
	return &itapi{
		Addr:   "http://test.srv.ittool.oa.com",
		RAddr:  "http://test.srv.ittool.oa.com",
		ModApi: "/api/DomainApI/UDNSSubmit",
		RApi:   "/api/dnsapi/GetIsDomainInfo",
	}
}

// newDefaultUNDS 默认UDNS配置
func newDefaultUNDS() *oudns {
	return &oudns{
		Addr:          "http://100.119.152.5/cgi-bin/udns/", //100.119.152.5代理9.22.61.22
		SystemID:      1389955267,
		Proposer:      "hermanzeng",
		AddApi:        "add_domain",
		AddZoneApi:    "add_zone_mnop",
		ModApi:        "mod_domain",
		DelApi:        "delete_domain",
		QueryApi:      "get_domain_conf",
		BatchQueryApi: "get_domain_conf_batch",
		BatchModApi:   "mod_domain_batch",
	}
}

// app 应用信息
type app struct {
	Name          string `toml:"name" env:"UDNS_APP_NAME"`
	Host          string `toml:"host" env:"UDNS_APP_HOST"`
	Port          string `toml:"port" env:"UDNS_APP_PORT"`
	Key           string `toml:"key" env:"UDNS_APP_KEY"`
	Mode          string `toml:"mode" env:"UDNS_APP_MODE"`
	GoroutineSize int    `toml:"goroutine_size" env:"UDNS_APP_GOROUTINE_SIZE"`
	ChanSize      int    `toml:"chan_size" env:"UDNS_APP_CHAN_SIZE"`
}

// Addr 获取地址
func (a *app) Addr() string {
	return a.Host + ":" + a.Port
}

// newDefaultAPP 默认APP
func newDefaultAPP() *app {
	return &app{
		Name:          "udns-api",
		Host:          "0.0.0.0",
		Port:          "8070",
		Key:           "default",
		Mode:          "dev",
		GoroutineSize: 20,
		ChanSize:      500,
	}
}

// log 日志
type log struct {
	Level       string `toml:"level" env:"UDNS_LOG_LEVEL"`
	PathDir     string `toml:"path_dir" env:"UDNS_LOG_PATH"`
	Format      string `toml:"format" env:"UDNS_LOG_FORMAT"`
	To          string `toml:"to" env:"UDNS_LOG_TO"`
	LogSDKTopic string `toml:"log_sdk_topic" env:"UDNS_LOG_SDK_TOPIC"`
	ZhiyanHost  string `toml:"zhiyan_host" env:"PERM_ZHIYAN_HOST"`
}

// newDefaultLog 默认日志
func newDefaultLog() *log {
	return &log{
		Level:       "debug",
		PathDir:     "logs",
		Format:      "json",
		To:          "stdout",
		LogSDKTopic: "fb-c9e5abe97fafeee",
		ZhiyanHost:  "",
	}
}

// flyPigeon 飞鸽
type flyPigeon struct {
	Addr         string `toml:"addr" env:"UDNS_FLYPIGEON_ADDR"`
	HMacID       string `toml:"chat_id" env:"UDNS_FLYPIGEON_HMAC_ID"`
	HMacSecret   string `toml:"bot_key" env:"UDNS_FLYPIGEON_HMAC_SECRET"`
	CorpID       string `toml:"chat_id" env:"UDNS_FLYPIGEON_CORP_ID"`
	CorpSecret   string `toml:"bot_key" env:"UDNS_FLYPIGEON_CORP_SECRET"`
	AppKey       string `toml:"app_key" env:"UDNS_FLYPIGEON_APPKEY"`
	SysID        string `toml:"sys_id" env:"UDNS_FLYPIGEON_SYSID"`
	MailSendFrom string `toml:"mail_send_from" env:"UDNS_FLYPIGEON_MAIL_SEND_FROM"`
}

// newDefaultFlyPigeon 飞鸽返回默认配置
func newDefaultFlyPigeon() *flyPigeon {
	return &flyPigeon{
		Addr:         "http://nops.tencent-cloud.com",
		HMacID:       "udns",
		HMacSecret:   "z83YyC5fGhkUN0oFaN6AwMMz5r2JwBSQ",
		CorpID:       "wxab249edd27d57738",
		CorpSecret:   "0Tf4fZTKYIewKK2MpJrOF870sMGha_GqXDG7AZVg0TI",
		AppKey:       "ad7eb985262242a68034d8355fd982ad",
		SysID:        "26070",
		MailSendFrom: "UDNS@tencent.com",
	}
}

// tof 部门信息
type tof struct {
	Host       string `toml:"host" env:"UDNS_TOF_HOST"`
	PaasID     string `toml:"paas_id" env:"UDNS_TOF_PAAS_ID"`
	PaasToken  string `toml:"paas_token" env:"UDNS_TOF_PAAS_TOKEN"`
	GSLBHost   string `toml:"gslb_host" env:"UDNS_GLSB_HOST"`
	SniperHost string `toml:"sniper_host" env:"UDNS_SNIPER_HOST"`
}

// newDefaultTOF 默认Tof
func newDefaultTOF() *tof {
	return &tof{
		Host:       "http://devnet.rio.tencent.com",
		PaasID:     "",
		PaasToken:  "",
		GSLBHost:   "http://gslb4.oa.com",
		SniperHost: "http://sniper.woa.com",
	}
}

// mongodb 数据库链接
type mongodb struct {
	Endpoints []string `toml:"endpoints" env:"UDNS_MONGO_ENDPOINTS" envSeparator:","`
	UserName  string   `toml:"username" env:"UDNS_MONGO_USERNAME"`
	Password  string   `toml:"password" env:"UDNS_MONGO_PASSWORD"`
	Database  string   `toml:"database" env:"UDNS_MONGO_DATABASE"`
}

// newDefaultMongoDB 默认mongodb
func newDefaultMongoDB() *mongodb {
	return &mongodb{
		Database:  "udns",
		Endpoints: []string{"127.0.0.1:27017"},
	}
}

// Client 获取一个全局的mongodb客户端连接
func (m *mongodb) Client() *mongo.Client {
	if mgoclient == nil {
		panic("please load mongo client first")
	}

	return mgoclient
}

// GetDB 获取数据库链接
func (m *mongodb) GetDB() *mongo.Database {
	return m.Client().Database(m.Database)
}

// getClient 数据库客户端
func (m *mongodb) getClient() (*mongo.Client, error) {
	opts := options.Client()

	cred := options.Credential{
		AuthSource: m.Database,
	}
	// 不属于dev环境的时候，默认使用云上mongo，source使用admin
	if C().App.Mode != "dev" {
		cred.AuthSource = mongoDefaultSource
	}

	if m.UserName != "" && m.Password != "" {
		cred.Username = m.UserName
		cred.Password = m.Password
		cred.PasswordSet = true
		opts.SetAuth(cred)
	}
	opts.SetHosts(m.Endpoints)
	opts.SetConnectTimeout(5 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("new mongodb client error, %s", err)
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("ping mongodb server(%s) error, %s", m.Endpoints, err)
	}

	return client, nil
}

// newDefaultCache 默认缓存配置
func newDefaultCache() *_cache {
	return &_cache{
		Type:    "memory",
		IsCache: true,
		Memory:  memory.NewDefaultConfig(),
		Redis:   redis.NewDefaultConfig(),
	}
}

// _cache 缓存配置
type _cache struct {
	Type    string         `toml:"type" json:"type" yaml:"type" env:"UDNS_CACHE_TYPE"`
	IsCache bool           `toml:"is_cache" json:"is_cache" yaml:"is_cache" env:"UDNS_CACHE_IS_CACHE"`
	Memory  *memory.Config `toml:"memory" json:"memory" yaml:"memory"`
	Redis   *redis.Config  `toml:"redis" json:"redis" yaml:"redis"`
}
