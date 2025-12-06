// Package auth 鉴权模块
package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wuzhengjerry/udns-api/conf"
	"github.com/wuzhengjerry/udns-api/frame/konghmac"
	"github.com/wuzhengjerry/udns-api/tool"
)

// AuthType 授权类型
type AuthType string

const (
	UserType   AuthType = "user"
	ModuleType AuthType = "module"
	ClientType AuthType = "client"
)

// AuthResponse 鉴权回包
type AuthResponse struct {
	Code int32       `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// AuthParam 权限校验参数，兼容多种校验字段
type AuthParam struct {
	AuthType      AuthType `json:"auth_type" binding:"required"`
	Subject       string   `json:"subject" binding:"required"`
	PermKey       string   `json:"perm_key" binding:"required"`
	Method        string   `json:"method" binding:"required"`
	BusinessID    string   `json:"business_id" binding:"required"`
	TimeStamp     string   `json:"time_stamp"`
	Authorization string   `json:"authorization"`
	Signature     string   `json:"signature"`
	StaffID       string   `json:"staff_id"`
	XExtData      string   `json:"x_ext_data"`
	XRioSEQ       string   `json:"x_rio_seq"`
	Digest        string   `json:"digest"`
	Date          string   `json:"date"`
	Body          []byte   `json:"body"`
}

// GetEnforceList 获取列表
func (a *AuthParam) GetEnforceList() []string {
	return []string{a.Subject, a.PermKey, a.Method, a.BusinessID}
}

// AuthUserSignature type为user时，使用iGate鉴权方式验证
func (a *AuthParam) AuthUserSignature() (ok bool, err error) {
	token := conf.C().Nops.IGateToken
	//timeStampStr := strconv.Itoa(a.TimeStamp)
	signStr := a.TimeStamp + token + a.XRioSEQ + "," + a.StaffID + "," + a.Subject + "," + a.XExtData + a.TimeStamp
	localSignature := strings.ToUpper(tool.GetSHA256HashCode([]byte(signStr)))
	// math.Abs(strconv.Itoa(a.TimeStamp)-float64(time.Now().Unix())) > 180
	if localSignature != a.Signature {
		return false, errors.New("bad signature or timestamp")
	}
	return true, nil
}

// AuthHmacSignature AuthSignature type为module或者alient的时候，使用hama验证
func (a *AuthParam) AuthHmacSignature(searet string) (ok bool, err error) {
	// 校验bodyDigest
	bodyDigest := konghmac.Sha256DigestBase64(a.Body)
	if a.Digest != bodyDigest {
		return
	}

	// 拼装待签名的数据
	strToSign := fmt.Sprintf("date: %s\ndigest: %s", a.Date, bodyDigest)

	//生成签名
	signature := konghmac.HmacSha256Base64(searet, strToSign)

	if signature != a.Signature {
		return
	}
	return true, nil
}
