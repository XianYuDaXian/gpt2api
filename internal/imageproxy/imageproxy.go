package imageproxy

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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

// CachePaths 返回缓存图片和元数据文件路径。
func CachePaths(cacheDir, taskID string, idx int) (imgPath, metaPath string) {
	base := filepath.Join(strings.TrimSpace(cacheDir), taskID, fmt.Sprintf("%d", idx))
	return base + ".bin", base + ".meta.json"
}

type cacheMeta struct {
	ContentType string    `json:"content_type"`
	StoredAt    time.Time `json:"stored_at"`
}

// LoadCache 从磁盘读取缓存图片。
func LoadCache(cacheDir, taskID string, idx int) ([]byte, string, bool, error) {
	imgPath, metaPath := CachePaths(cacheDir, taskID, idx)
	body, err := os.ReadFile(imgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", false, nil
		}
		return nil, "", false, err
	}
	ct := ""
	if metaBytes, err := os.ReadFile(metaPath); err == nil {
		var meta cacheMeta
		if json.Unmarshal(metaBytes, &meta) == nil {
			ct = meta.ContentType
		}
	}
	if ct == "" {
		ct = "image/png"
	}
	return body, ct, true, nil
}

// StoreCache 写入磁盘缓存图片。
func StoreCache(cacheDir, taskID string, idx int, body []byte, contentType string) error {
	if strings.TrimSpace(cacheDir) == "" || len(body) == 0 {
		return nil
	}
	imgPath, metaPath := CachePaths(cacheDir, taskID, idx)
	if err := os.MkdirAll(filepath.Dir(imgPath), 0o750); err != nil {
		return err
	}
	tmpImg := imgPath + ".tmp"
	if err := os.WriteFile(tmpImg, body, 0o640); err != nil {
		return err
	}
	_ = os.Remove(imgPath)
	if err := os.Rename(tmpImg, imgPath); err != nil {
		_ = os.Remove(tmpImg)
		return err
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "image/png"
	}
	meta := cacheMeta{ContentType: contentType, StoredAt: time.Now()}
	metaBytes, _ := json.Marshal(meta)
	tmpMeta := metaPath + ".tmp"
	if err := os.WriteFile(tmpMeta, metaBytes, 0o640); err != nil {
		return err
	}
	_ = os.Remove(metaPath)
	if err := os.Rename(tmpMeta, metaPath); err != nil {
		_ = os.Remove(tmpMeta)
		return err
	}
	return nil
}
