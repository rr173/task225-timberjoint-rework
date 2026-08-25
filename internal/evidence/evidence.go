// Package evidence 承载现场照片证据的业务规则：校验与幂等哈希。
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"task225-timberjoint/internal/model"
)

// ValidateEvidence 校验照片证据：文件名非空、说明非空、拍摄时间不晚于当前。
func ValidateEvidence(filename, caption string, takenAt time.Time) error {
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("%w: filename required", model.ErrInvalidInput)
	}
	if strings.TrimSpace(caption) == "" {
		return fmt.Errorf("%w: caption required", model.ErrInvalidInput)
	}
	if takenAt.IsZero() || takenAt.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("%w: invalid taken_at time", model.ErrInvalidInput)
	}
	return nil
}

// ContentHash 计算照片内容的幂等哈希（文件名 + 说明 + 拍摄时间）。
// 同一照片重复上传将得到相同哈希，用于幂等去重。
func ContentHash(filename, caption string, takenAt time.Time) string {
	h := sha256.Sum256([]byte(strings.Join([]string{
		filename,
		caption,
		takenAt.UTC().Format(time.RFC3339),
	}, "\x00")))
	return hex.EncodeToString(h[:])
}
