package geometry

import "math"

// DirectionResult 是节点构件方向冲突检测的结果。
type DirectionResult struct {
	Spread       float64 // 方向散布：所有构件方向两两夹角最大值（度）
	PrimaryAxis  Vec3    // 节点主方向（以最长构件的方向为参考）
	Conflict     bool    // 是否存在方向冲突（散布 > 容差角）
	MaxPair      [2]string // 夹角最大的两个构件编号
}

// MemberAxis 是参与方向检测的一个构件方向样本。
type MemberAxis struct {
	Label string // 构件编号
	Dir   Vec3   // 构件方向单位向量（终点-起点）
}

// AnalyzeDirection 计算一组构件方向向量的散布与冲突：
//  1. 以最长（范数最大）构件的方向为节点主方向参考；
//  2. 计算所有构件方向两两夹角，取最大值为散布角；
//  3. 散布角超过容差 tolDeg 即判为方向冲突。
func AnalyzeDirection(axes []MemberAxis, tolDeg float64) DirectionResult {
	res := DirectionResult{MaxPair: [2]string{}}
	if len(axes) < 2 {
		return res
	}
	// 主方向：取范数最大的构件方向。
	bestNorm := -1.0
	for _, a := range axes {
		if n := a.Dir.Norm(); n > bestNorm {
			bestNorm = n
			res.PrimaryAxis = a.Dir.Unit()
		}
	}
	for i := 0; i < len(axes); i++ {
		for j := i + 1; j < len(axes); j++ {
			ang := AngleDeg(axes[i].Dir, axes[j].Dir)
			if ang > res.Spread {
				res.Spread = ang
				res.MaxPair = [2]string{axes[i].Label, axes[j].Label}
			}
		}
	}
	res.Conflict = res.Spread > tolDeg
	return res
}

// ToleranceDeg 归一化角度容差：负值/零值回退到默认 10 度。
func ToleranceDeg(tol float64) float64 {
	if tol <= 0 {
		return 10.0
	}
	return tol
}

// Radians 角度转弧度。
func Radians(deg float64) float64 { return deg * math.Pi / 180 }

// Degrees 弧度转角度。
func Degrees(rad float64) float64 { return rad * 180 / math.Pi }
