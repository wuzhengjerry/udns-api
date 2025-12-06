// Package tool 工具模块
package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	// CacheTTL
	CacheTTL = 360 * time.Second
	// BusinessTreeCacheKey 默认cacheKey
	BusinessTreeCacheKey = "business_tree"
)

// BusinessTreeResponse 组织架构数
type BusinessTreeResponse struct {
	RetCode string            `json:"ret_code"`
	Data    map[string]string `json:"data"`
	Msg     string            `json:"msg"`
}

// MakeUUID use to make bearer random token
func MakeUUID(lenth int) string {
	charlist := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	password := make([]string, lenth)
	rand.Seed(time.Now().UnixNano() + int64(lenth))
	for i := 0; i < lenth; i++ {
		rn := rand.Intn(len(charlist))
		w := charlist[rn : rn+1]
		password = append(password, w)
	}

	return strings.Join(password, "")
}

// GetNowTimeStamp 获取当前unix时间戳
func GetNowTimeStamp() int64 {
	return time.Now().Unix()
}

// SplitAuthorization 鉴权相关
func SplitAuthorization(authorization string) (subject, signature string, err error) {
	// Authorization: `hmac username="%s", algorithm="hmac-sha256", headers="date digest", signature="%s"`
	list := strings.Split(strings.TrimSpace(authorization), ",")
	if len(list) != 3 {
		err = errors.New("bad authorization format")
		return
	}
	return list[0], list[3], nil
}

// GetSHA256HashCode 获取sha加密结果
func GetSHA256HashCode(message []byte) string {
	//方法一：
	//创建一个基于SHA256算法的hash.Hash接口的对象
	hash := sha256.New()
	//输入数据
	hash.Write(message)
	//计算哈希值
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashCode := hex.EncodeToString(bytes)
	//返回哈希值
	return hashCode

	//方法二：
	//bytes2:=sha256.Sum256(message)//计算哈希值，返回一个长度为32的数组
	//hashcode2:=hex.EncodeToString(bytes2[:])//将数组转换成切片，转换成16进制，返回字符串
	//return hashcode2
}

// GetExternalIP 获取外部ip
func GetExternalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("get external ip error:", err)
		return "1.1.1.1"
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return (ipnet.IP.String())
			}
		}
	}
	return "1.1.1.1"
}

// IsIPv4 判断是否ipv4
func IsIPv4(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ".")
}

// IsIPv6 判断是否ipv6
func IsIPv6(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ":")
}

// IsValidDomain 是否合法域名
func IsValidDomain(domain string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9]+([\\-\\.][a-zA-Z0-9]+)*\\.[a-zA-Z]{2,}$")
	return r.MatchString(domain)
}
