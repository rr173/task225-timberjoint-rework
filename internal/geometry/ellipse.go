package geometry

import "math"

// ErrorEllipsoid 表示测点的误差椭球：三正交半轴 + 主轴方位角。
// 半轴（semi-major/semi-minor/vertical）反映该测点在各方向上的不确定度；
// 主轴方位角（orientation azimuth，度）描述长半轴在水平面内的朝向。
type ErrorEllipsoid struct {
	SemiMajor  float64 // 长半轴
	SemiMinor  float64 // 短半轴
	Vertical   float64 // 垂向半轴
	Orientation float64 // 主轴方位角（度）
}

// NewErrorEllipsoid 构造误差椭球，负半轴归一为 0。
func NewErrorEllipsoid(major, minor, vertical, orientation float64) ErrorEllipsoid {
	return ErrorEllipsoid{
		SemiMajor:   math.Max(0, major),
		SemiMinor:   math.Max(0, minor),
		Vertical:    math.Max(0, vertical),
		Orientation: orientation,
	}
}

// MaxSemiAxis 返回三半轴最大值，作为该测点的最大单向不确定度。
func (e ErrorEllipsoid) MaxSemiAxis() float64 {
	return math.Max(e.SemiMajor, math.Max(e.SemiMinor, e.Vertical))
}

// HorizontalSemiAxes 返回水平面内两半轴（长、短）。
func (e ErrorEllipsoid) HorizontalSemiAxes() (float64, float64) {
	return e.SemiMajor, e.SemiMinor
}

// EffectiveRadius 返回用于闭合判定的等效不确定半径：
// 取水平两半轴的均方根与垂向半轴的组合，近似该测点的空间不确定球半径。
func (e ErrorEllipsoid) EffectiveRadius() float64 {
	horizontal := math.Hypot(e.SemiMajor, e.SemiMinor)
	return math.Hypot(horizontal, e.Vertical)
}

// WithinBudget 判断闭合残差是否落在测点误差椭球的允许预算内。
// 采用 RSS 合成所有测点等效半径，再与一个放宽系数比较，判断散布是否
// 可用测量不确定度解释（而非结构缺陷）。
func (e ErrorEllipsoid) WithinBudget(residual float64, radii []float64, relax float64) bool {
	if len(radii) == 0 {
		return false
	}
	var sum float64
	for _, r := range radii {
		sum += r * r
	}
	budget := math.Sqrt(sum) * relax
	return residual <= budget
}
