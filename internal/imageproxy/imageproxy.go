package imageproxy

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// ImageProxyTTL 单条签名 URL 的默认有效期。
const ImageProxyTTL = 24 * time.Hour

var secret []byte

func init() {
	secret = make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		for i := range secret {
			secret[i] = byte(i*31 + 7)
		}
	}
}

// BuildURL 生成图片代理 URL。
// baseURL 为空时返回相对路径,有值时返回绝对地址。
func BuildURL(baseURL, taskID string, idx int, ttl time.Duration) string {
	if ttl <= 0 {
		ttl = ImageProxyTTL
	}
	expMs := time.Now().Add(ttl).UnixMilli()
	sig := Sign(taskID, idx, expMs)
	rel := fmt.Sprintf("/p/img/%s/%d?exp=%d&sig=%s", taskID, idx, expMs, sig)
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return rel
	}
	return baseURL + rel
}

// Sign 生成签名。
func Sign(taskID string, idx int, expMs int64) string {
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "%s|%d|%d", taskID, idx, expMs)
	return hex.EncodeToString(mac.Sum(nil))[:24]
}

// VerifySig 校验签名与过期时间。
func VerifySig(taskID string, idx int, expMs int64, sig string) bool {
	if expMs < time.Now().UnixMilli() {
		return false
	}
	want := Sign(taskID, idx, expMs)
	return hmac.Equal([]byte(sig), []byte(want))
}
