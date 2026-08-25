// Package version 承载节点版本（不可变快照）的业务规则。
package version

import (
	"fmt"
	"strings"

	"task225-timberjoint/internal/model"
)

// 版本状态机允许的流转表。
var versionTransitions = map[string][]string{
	model.VersionDraft:  {model.VersionShared, model.VersionFrozen},
	model.VersionShared: {model.VersionFrozen},
	model.VersionFrozen: {model.VersionSuperseded},
}

// CanTransition 判断版本状态机是否允许 from -> to。
func CanTransition(from, to string) bool {
	for _, t := range versionTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// ValidateReason 校验发布/冻结原因：冻结时必须给出原因。
func ValidateReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("%w: reason required", model.ErrInvalidInput)
	}
	return nil
}

// NextVersionNo 计算同节点下一版本号（当前最大版本 + 1）。
func NextVersionNo(currentMax int) int { return currentMax + 1 }

// IsFrozenOrSuperseded 判断版本是否已冻结或已被替代（不可再编辑）。
func IsFrozenOrSuperseded(status string) bool {
	return status == model.VersionFrozen || status == model.VersionSuperseded
}
