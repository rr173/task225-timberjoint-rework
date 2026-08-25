package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// fingerprint 计算幂等指纹：把若干字段以 \x00 连接后取 SHA-256 前 16 字节。
func fingerprint(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h[:16])
}

// coordKey 生成坐标字符串（用于指纹）。
func coordKey(x, y, z float64) string {
	return fmt.Sprintf("%.6f,%.6f,%.6f", x, y, z)
}
