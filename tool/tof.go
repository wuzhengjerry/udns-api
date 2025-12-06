package tool

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wuzhengjerry/udns-api/conf"
)

// UserBaseInfo 用户信息
type UserBaseInfo struct {
	ID                       int    `json:"Id"`
	EnglishName              string `json:"EnglishName"`
	DepartmentName           string `json:"DepartmentName"`
	DepartmentLocationString string `json:"DepartmentLocationString"`
	DepartmentFullName       string `json:"DepartmentFullName"`
}

// TOFResponse TOF 返回
type TOFResponse struct {
	Ret        int           `json:"Ret"`
	ErrCode    int           `json:"ErrCode"`
	ErrMsg     string        `json:"ErrMsg"`
	StackTrace string        `json:"StackTrace"`
	Data       *UserBaseInfo `json:"Data"`
}

// getUserBaseInfo 获取用户基础信息
func getUserBaseInfo(user string) (res *TOFResponse, err error) {
	paasId := conf.C().TOF.PaasID                                                            // 在TOF门户注册获得的应用id
	paasToken := conf.C().TOF.PaasToken                                                      // 在TOF门户注册获得的签名密钥
	server := conf.C().TOF.Host                                                              // 智能网关在DevCloud区的接入点域名
	path := fmt.Sprintf("/ebus/tof4_org/api/v1/staff/getstaffbaseinfo?id=&engName=%s", user) // 在TOF门户订阅接口成功后，获得的接口path

	//params := "" // 接口入参,建议以结构体的形式而不是字符串的形式入参，此处字符串只是示例。具体参考接口文档。
	timestamp := fmt.Sprintf("%d", time.Now().Unix()) // 生成时间戳，注意服务器的时间与标准时间差不能大于180秒
	rand.Seed(time.Now().Unix())
	r := rand.New(rand.NewSource(time.Now().Unix()))
	nonce := strconv.Itoa(r.Intn(4096)) // 随机字符串，十分钟内不重复即可
	signStr := fmt.Sprintf("%s%s%s%s", timestamp, paasToken, nonce, timestamp)
	sign := fmt.Sprintf("%X", sha256.Sum256([]byte(signStr))) // 输出大写的结果
	req, err := http.NewRequest("GET", server+path, nil)
	if err != nil {
		return nil, err
	}

	var httpClient = &http.Client{}
	// 设置鉴权参数，如果此参数设置错误，将触发“AGW.xxxxx”类型的错误，详见3.4章节
	req.Header.Add("x-rio-paasid", paasId)
	req.Header.Add("x-rio-nonce", nonce)
	req.Header.Add("x-rio-timestamp", timestamp)
	req.Header.Add("x-rio-signature", sign)

	// 这里主要展示http head的构造，省略了http body的构造。
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("request tof error: %s", string(content))
		fmt.Println(errStr)
		return nil, errors.New(errStr)
	}
	err = json.Unmarshal(content, &res)
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		fmt.Println("bad tof response:", res.ErrMsg)
	}
	fmt.Println(res.Data.DepartmentFullName)
	return
}

// IsCSIGUser 判断用户是否属于 CSIG云与智慧产业事业群
func IsCSIGUser(user string) (bool, error) {
	userIngo, err := getUserBaseInfo(user)
	if err != nil {
		return false, err
	}
	return strings.Contains(userIngo.Data.DepartmentFullName, "CSIG云与智慧产业事业群"), nil
}
