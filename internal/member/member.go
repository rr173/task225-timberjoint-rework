// Package member 承载木构构件（柱/梁/枋/斗/拱）与榫卯描述的业务规则。
package member

import (
	"fmt"
	"math"
	"strings"

	"task225-timberjoint/internal/geometry"
	"task225-timberjoint/internal/model"
)

// 允许的构件类型白名单。
var validMemberTypes = map[string]bool{
	"柱": true, "梁": true, "枋": true, "斗": true, "拱": true, "其他": true,
}

// ValidateMember 校验构件：编号非空、端点坐标有限、两端点不重合、
// 类型在白名单内。返回构件方向样本（用于方向冲突判定）。
func ValidateMember(no, memberType string, sx, sy, sz, ex, ey, ez float64) (geometry.MemberAxis, error) {
	if strings.TrimSpace(no) == "" {
		return geometry.MemberAxis{}, fmt.Errorf("%w: member number required", model.ErrInvalidInput)
	}
	if !validMemberTypes[strings.TrimSpace(memberType)] {
		return geometry.MemberAxis{}, fmt.Errorf("%w: unknown member type %q", model.ErrInvalidInput, memberType)
	}
	for _, v := range []float64{sx, sy, sz, ex, ey, ez} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return geometry.MemberAxis{}, fmt.Errorf("%w: endpoint coordinate must be finite", model.ErrInvalidInput)
		}
	}
	start := geometry.Vec3{X: sx, Y: sy, Z: sz}
	end := geometry.Vec3{X: ex, Y: ey, Z: ez}
	dir := end.Sub(start)
	if dir.Norm() < 1e-12 {
		return geometry.MemberAxis{}, fmt.Errorf("%w: member endpoints coincide", model.ErrInvalidInput)
	}
	return geometry.MemberAxis{Label: no, Dir: dir}, nil
}

// MemberDirection 返回构件的方向单位向量（用于报告/序列化）。
func MemberDirection(sx, sy, sz, ex, ey, ez float64) geometry.Vec3 {
	return geometry.Vec3{X: ex - sx, Y: ey - sy, Z: ez - sz}.Unit()
}

// MemberMissingOrConflict 判断构件状态是否阻碍节点闭合（缺失/方向冲突）。
func MemberMissingOrConflict(status string) bool {
	return status == model.MemberMissing || status == model.MemberDirectionConflict
}
