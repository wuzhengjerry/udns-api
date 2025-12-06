package auth

import (
	"fmt"

	"github.com/wuzhengjerry/udns-api/frame/konghmac"
)

// GenHmacSignature AuthSignature type为module或者client的时候，使用hamc验证
func (a *AuthParam) GenHmacSignature(secret string) (ok bool, err error) {
	// 校验bodyDigest
	bodyDigest := konghmac.Sha256DigestBase64(a.Body)
	if a.Digest != bodyDigest {
		return
	}

	// 拼装待签名的数据
	strToSign := fmt.Sprintf("date: %s\ndigest: %s", a.Date, bodyDigest)

	//生成签名
	signature := konghmac.HmacSha256Base64(secret, strToSign)

	if signature != a.Signature {
		return
	}
	return true, nil
}
