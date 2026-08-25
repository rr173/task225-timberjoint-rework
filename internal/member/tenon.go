package member

import (
	"encoding/json"
	"fmt"

	"task225-timberjoint/internal/model"
)

// TenonSpec 描述一个榫卯节点：榫头/卯口的类型与尺寸。
// 用于把构件端部的榫卯连接方式结构化，供闭合与方向复核引用。
type TenonSpec struct {
	TenonType string  `json:"tenon_type"` // 榫头类型：直榫/燕尾榫/馒头榫/半榫等
	Mortise   string  `json:"mortise"`    // 对应卯口类型
	Width     float64 `json:"width"`      // 榫宽（单位同批次）
	Depth     float64 `json:"depth"`      // 榫深（单位同批次）
	Note      string  `json:"note"`       // 备注
}

// ParseTenonDesc 解析构件的榫卯描述 JSON；空串返回空描述。
func ParseTenonDesc(raw string) (TenonSpec, error) {
	if raw == "" {
		return TenonSpec{}, nil
	}
	var spec TenonSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return TenonSpec{}, fmt.Errorf("%w: invalid tenon description: %v", model.ErrInvalidInput, err)
	}
	return spec, nil
}

// ValidateTenonSpec 校验榫卯描述：榫宽/深非负，类型非空。
func ValidateTenonSpec(spec TenonSpec) error {
	if spec.TenonType == "" && spec.Mortise == "" {
		return nil // 允许空描述（构件仅登记端点，无榫卯信息）
	}
	if spec.Width < 0 || spec.Depth < 0 {
		return fmt.Errorf("%w: tenon width/depth must be non-negative", model.ErrInvalidInput)
	}
	return nil
}
