package geometry

import "math"

// ClosureResult 是节点几何闭合检测的结果。
type ClosureResult struct {
	Centroid      Vec3    // 参与闭合点的质心
	Residual      float64 // 最大残差：点到质心的最大距离
	PointCount    int     // 参与闭合的点数
	Closed        bool    // 是否闭合（残差 <= 容差）
	ContactErrors int     // 被剔除的接触异常测点数
	Gaps          int     // 被剔除的缺口测点数
}

// PointSample 是参与闭合判定的一个空间采样点，携带其编号与误差半径。
type PointSample struct {
	Label   string  // 测点/端点编号（用于报告）
	Pos     Vec3    // 三维坐标
	Radius  float64 // 误差椭球等效半径（单位同坐标）
	Exclude bool    // 是否剔除（接触异常/缺口）
}

// AnalyzeClosure 计算一组采样点的闭合性：
//  1. 剔除被标记为接触异常/缺口的点（保留原始数据，但不参与闭合）；
//  2. 计算有效点质心；
//  3. 计算最大残差（点到质心最大距离）；
//  4. 以容差判断闭合：残差 <= 容差即闭合。
//
// 容差 tolerance 由批次声明；有效点少于 2 个时视为无法闭合。
func AnalyzeClosure(samples []PointSample, tolerance float64) ClosureResult {
	res := ClosureResult{}
	var valid []PointSample
	for _, s := range samples {
		if s.Exclude {
			res.ContactErrors++
			continue
		}
		valid = append(valid, s)
	}
	res.PointCount = len(valid)
	if len(valid) < 2 {
		res.Closed = false
		return res
	}
	pts := make([]Vec3, 0, len(valid))
	for _, s := range valid {
		pts = append(pts, s.Pos)
	}
	res.Centroid = Centroid(pts)
	for _, s := range valid {
		if d := s.Pos.Distance(res.Centroid); d > res.Residual {
			res.Residual = d
		}
	}
	res.Closed = res.Residual <= tolerance
	return res
}

// ClosureBudgetResidual 计算考虑测点误差椭球后的预算残差：
// 若最大残差可用各点不确定度 RSS 合成解释（放宽 relax 倍），
// 则即便略超几何容差也可判定为“误差预算内闭合”。
func ClosureBudgetResidual(samples []PointSample, relax float64) float64 {
	var sum float64
	n := 0
	for _, s := range samples {
		if s.Exclude {
			continue
		}
		sum += s.Radius * s.Radius
		n++
	}
	if n == 0 {
		return 0
	}
	return math.Sqrt(sum) * relax
}
