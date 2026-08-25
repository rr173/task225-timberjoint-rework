package joint

import (
	"encoding/json"
	"fmt"

	"task225-timberjoint/internal/geometry"
	"task225-timberjoint/internal/model"
)

// PointInfo 是节点复核的测点输入（由 service 从存储层组装）。
type PointInfo struct {
	No       string         // 测点编号
	Pos      geometry.Vec3  // 三维坐标
	Radius   float64        // 误差椭球等效半径
	Excluded bool           // 接触异常/缺口，不参与闭合
}

// MemberInfo 是节点复核的构件输入。
type MemberInfo struct {
	No      string        // 构件编号
	Dir     geometry.Vec3 // 构件方向向量
	Missing bool          // 缺失构件，不参与方向判定
}

// CheckReport 是节点几何复核的完整报告。
type CheckReport struct {
	Closed            bool     `json:"closed"`             // 是否闭合
	ClosureResidual   float64  `json:"closure_residual"`   // 闭合残差（最大散布距离）
	DirectionSpread   float64  `json:"direction_spread"`   // 方向散布（最大夹角，度）
	DirectionConflict bool     `json:"direction_conflict"` // 是否存在方向冲突
	PrimaryAxis       geometry.Vec3 `json:"primary_axis"`  // 节点主方向
	NumberingOK       bool     `json:"numbering_ok"`       // 编号一致性是否通过
	Duplicates        []string `json:"duplicates"`         // 重复编号
	Malformed         []string `json:"malformed"`          // 格式非法编号
	ContactErrors     int      `json:"contact_errors"`     // 接触异常测点数
	Summary           string   `json:"summary"`            // 结论摘要
}

// CheckNode 执行节点几何复核：闭合检测 + 方向冲突 + 编号一致性。
func CheckNode(points []PointInfo, members []MemberInfo, closureTol, directionTol float64) CheckReport {
	rep := CheckReport{}

	// 1. 编号一致性（测点编号 + 构件编号）。
	allNos := make([]string, 0, len(points)+len(members))
	for _, p := range points {
		allNos = append(allNos, p.No)
	}
	for _, m := range members {
		allNos = append(allNos, m.No)
	}
	numRes := geometry.CheckNumbering(allNos)
	rep.NumberingOK = numRes.OK
	rep.Duplicates = numRes.Duplicates
	rep.Malformed = numRes.Malformed

	// 2. 闭合检测。
	samples := make([]geometry.PointSample, 0, len(points))
	for _, p := range points {
		samples = append(samples, geometry.PointSample{
			Label:   p.No,
			Pos:     p.Pos,
			Radius:  p.Radius,
			Exclude: p.Excluded,
		})
	}
	closure := geometry.AnalyzeClosure(samples, closureTol)
	rep.Closed = closure.Closed
	rep.ClosureResidual = closure.Residual
	rep.ContactErrors = closure.ContactErrors

	// 3. 方向冲突。
	axes := make([]geometry.MemberAxis, 0, len(members))
	for _, m := range members {
		if m.Missing {
			continue
		}
		axes = append(axes, geometry.MemberAxis{Label: m.No, Dir: m.Dir})
	}
	dirRes := geometry.AnalyzeDirection(axes, geometry.ToleranceDeg(directionTol))
	rep.DirectionSpread = dirRes.Spread
	rep.DirectionConflict = dirRes.Conflict
	rep.PrimaryAxis = dirRes.PrimaryAxis

	rep.Summary = summarize(rep)
	return rep
}

// summarize 生成结论摘要。
func summarize(rep CheckReport) string {
	switch {
	case !rep.NumberingOK:
		return fmt.Sprintf("编号冲突：重复=%v 非法=%v", rep.Duplicates, rep.Malformed)
	case rep.DirectionConflict:
		return fmt.Sprintf("方向冲突：散布 %.2f°", rep.DirectionSpread)
	case rep.Closed:
		return fmt.Sprintf("闭合：残差 %.4f", rep.ClosureResidual)
	default:
		return fmt.Sprintf("断裂：残差 %.4f 超容差", rep.ClosureResidual)
	}
}

// MarshalReport 序列化校验报告为 JSON 字符串。
func MarshalReport(rep CheckReport) string {
	b, _ := json.Marshal(rep)
	return string(b)
}

// UnmarshalReport 反序列化校验报告。
func UnmarshalReport(raw string) (CheckReport, error) {
	var rep CheckReport
	if err := json.Unmarshal([]byte(raw), &rep); err != nil {
		return CheckReport{}, fmt.Errorf("%w: invalid check report: %v", model.ErrInvalidInput, err)
	}
	return rep, nil
}
