package survey

import (
	"fmt"
	"math"
	"strings"

	"task225-timberjoint/internal/geometry"
	"task225-timberjoint/internal/model"
)

// ValidatePoint 校验测点：编号非空、坐标为有限值、误差椭球三半轴非负。
// 返回测点对应的几何采样点（用于闭合判定）。
func ValidatePoint(no string, x, y, z, semiMajor, semiMinor, vertical, orient float64) (geometry.PointSample, error) {
	if strings.TrimSpace(no) == "" {
		return geometry.PointSample{}, fmt.Errorf("%w: point number required", model.ErrInvalidInput)
	}
	for _, v := range []float64{x, y, z} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return geometry.PointSample{}, fmt.Errorf("%w: coordinate must be finite", model.ErrInvalidInput)
		}
	}
	for _, v := range []float64{semiMajor, semiMinor, vertical} {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return geometry.PointSample{}, fmt.Errorf("%w: error semi-axes must be non-negative finite", model.ErrInvalidInput)
		}
	}
	ell := geometry.NewErrorEllipsoid(semiMajor, semiMinor, vertical, orient)
	return geometry.PointSample{
		Label:  no,
		Pos:    geometry.Vec3{X: x, Y: y, Z: z},
		Radius: ell.EffectiveRadius(),
	}, nil
}

// PointExcluded 判断某状态的测点是否不参与闭合判定（接触异常/缺口保留原始数据）。
func PointExcluded(status string) bool {
	return status == model.PointContactError || status == model.PointGap
}

// DefaultPointStatus 返回新测点的初始状态。
func DefaultPointStatus() string { return model.PointPending }
