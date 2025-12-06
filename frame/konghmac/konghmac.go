// Package konghmac 签名模块
package konghmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// GetAuthHeader 获取请求头
func GetAuthHeader(username, secretkey string, body []byte) map[string]string {
	//生成body的sha256加密串
	bodyDigest := Sha256DigestBase64(body)

	gmTime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	//拼装待签名的数据
	strToSign := fmt.Sprintf("date: %s\ndigest: %s", gmTime, bodyDigest)

	//生成签名
	signature := HmacSha256Base64(secretkey, strToSign)

	authStr := fmt.Sprintf(`hmac username="%s", algorithm="hmac-sha256", headers="date digest", signature="%s"`,
		username, signature)
	return map[string]string{
		"Authorization": authStr,
		"Digest":        bodyDigest,
		"Date":          gmTime,
		//"Grant-Type":    grantType,
	}
}

// Sha256DigestBase64 加密
func Sha256DigestBase64(body []byte) string {
	sha := sha256.New()
	sha.Write(body)
	return fmt.Sprintf("SHA-256=%s", base64.StdEncoding.EncodeToString(sha.Sum(nil)))
}

// HmacSha256Base64 加密
func HmacSha256Base64(secretkey, strToSign string) string {
	h := hmac.New(sha256.New, []byte(secretkey))
	h.Write([]byte(strToSign))
	result := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(result)
}
