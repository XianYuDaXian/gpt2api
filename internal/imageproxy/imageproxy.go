package imageproxy

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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

// CachePaths 返回默认缓存图片和元数据文件路径。
func CachePaths(cacheDir, taskID string, idx int) (imgPath, metaPath string) {
	base := filepath.Join(strings.TrimSpace(cacheDir), taskID, fmt.Sprintf("%d", idx))
	return base + ".png", base + ".meta.json"
}

type cacheMeta struct {
	ContentType string    `json:"content_type"`
	StoredAt    time.Time `json:"stored_at"`
}

// LoadCache 从磁盘读取缓存图片。
func LoadCache(cacheDir, taskID string, idx int) ([]byte, string, bool, error) {
	ct := ""
	_, metaPath := CachePaths(cacheDir, taskID, idx)
	if metaBytes, err := os.ReadFile(metaPath); err == nil {
		var meta cacheMeta
		if json.Unmarshal(metaBytes, &meta) == nil {
			ct = meta.ContentType
		}
	}
	if ct == "" {
		ct = "image/png"
	}
	var lastErr error
	for _, imgPath := range cacheImageCandidates(cacheDir, taskID, idx, ct) {
		body, err := os.ReadFile(imgPath)
		if err == nil {
			return body, ct, true, nil
		}
		if !os.IsNotExist(err) {
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, "", false, lastErr
	}
	return nil, "", false, nil
}

func cacheImageCandidates(cacheDir, taskID string, idx int, contentType string) []string {
	base := filepath.Join(strings.TrimSpace(cacheDir), taskID, fmt.Sprintf("%d", idx))
	ext := imageExt(contentType)
	out := []string{base + ext}
	// 兼容旧版本缓存文件。
	if ext != ".bin" {
		out = append(out, base+".bin")
	}
	for _, fallback := range []string{".png", ".jpg", ".jpeg", ".webp", ".gif"} {
		p := base + fallback
		exists := false
		for _, cur := range out {
			if cur == p {
				exists = true
				break
			}
		}
		if !exists {
			out = append(out, p)
		}
	}
	return out
}

func cacheImagePath(cacheDir, taskID string, idx int, contentType string) string {
	base := filepath.Join(strings.TrimSpace(cacheDir), taskID, fmt.Sprintf("%d", idx))
	return base + imageExt(contentType)
}

func imageExt(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ct {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/png":
		return ".png"
	default:
		return ".png"
	}
}

func normalizeImageContentType(contentType string, body []byte) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ct {
	case "image/png", "image/jpeg", "image/webp", "image/gif":
		return ct
	case "image/jpg":
		return "image/jpeg"
	}
	if len(body) > 0 {
		sniff := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(body), ";")[0]))
		switch sniff {
		case "image/png", "image/jpeg", "image/webp", "image/gif":
			return sniff
		}
	}
	return "image/png"
}

// StoreCache 写入磁盘缓存图片。
func StoreCache(cacheDir, taskID string, idx int, body []byte, contentType string) error {
	if strings.TrimSpace(cacheDir) == "" || len(body) == 0 {
		return nil
	}
	contentType = normalizeImageContentType(contentType, body)
	imgPath := cacheImagePath(cacheDir, taskID, idx, contentType)
	_, metaPath := CachePaths(cacheDir, taskID, idx)
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
