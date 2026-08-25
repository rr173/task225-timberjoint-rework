package store

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// newID 生成带前缀的唯一 ID（前缀 + 随机 8 字节 hex + 纳秒时间戳 36 进制）。
func newID(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:]) + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// isUniqueViolation 判断 SQLite 唯一约束冲突错误（UNIQUE / PRIMARY KEY）。
// modernc.org/sqlite 返回的错误文本包含 "UNIQUE constraint failed"。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "constraint failed") ||
		strings.Contains(msg, "primary key")
}

// isFKViolation 判断外键约束冲突（引用不存在的批次/节点关系）。
func isFKViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key constraint failed") ||
		strings.Contains(msg, "foreign key")
}
