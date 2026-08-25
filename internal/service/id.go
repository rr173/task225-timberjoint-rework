package service

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// newServiceID 生成带前缀的唯一 ID（与 store 同风格，service 层自用）。
func newServiceID(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:]) + strconv.FormatInt(time.Now().UnixNano(), 36)
}
