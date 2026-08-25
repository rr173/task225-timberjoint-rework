// Package survey 承载测绘批次与测点的业务规则（不含持久化与 HTTP）。
package survey

import (
	"fmt"
	"strings"

	"task225-timberjoint/internal/model"
)

// 批次状态机允许的流转表。
var batchTransitions = map[string][]string{
	model.BatchCollecting: {model.BatchOrganizing},
	model.BatchOrganizing: {model.BatchReviewing},
	model.BatchReviewing:  {model.BatchPublished},
	model.BatchPublished:  {model.BatchArchived},
}

// ValidateCoordinateSystem 校验坐标系声明：非空且来自白名单（可扩展）。
func ValidateCoordinateSystem(cs string) error {
	if strings.TrimSpace(cs) == "" {
		return fmt.Errorf("%w: coordinate system must be declared", model.ErrInvalidInput)
	}
	return nil
}

// ValidateLengthUnit 校验长度单位：仅支持 m / mm。
func ValidateLengthUnit(unit string) error {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "m", "mm":
		return nil
	default:
		return fmt.Errorf("%w: unsupported length unit %q", model.ErrInvalidInput, unit)
	}
}

// ValidateTolerances 校验闭合容差与方向容差为正值。
func ValidateTolerances(closure float64, directionDeg float64) error {
	if closure <= 0 {
		return fmt.Errorf("%w: closure tolerance must be positive", model.ErrInvalidInput)
	}
	if directionDeg <= 0 || directionDeg > 180 {
		return fmt.Errorf("%w: direction tolerance must be in (0,180]", model.ErrInvalidInput)
	}
	return nil
}

// CanTransition 判断批次状态机是否允许 from -> to。
func CanTransition(from, to string) bool {
	for _, t := range batchTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// IsMutable 判断批次当前状态是否允许写入测点/构件/关系。
func IsMutable(status string) bool {
	return status == model.BatchCollecting || status == model.BatchOrganizing || status == model.BatchReviewing
}

// NextBatchState 返回批次的下一状态；无后继返回空串。
func NextBatchState(from string) string {
	next := batchTransitions[from]
	if len(next) == 0 {
		return ""
	}
	return next[0]
}
