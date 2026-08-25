// Package geometry 承载木构节点复核的几何计算：三维向量运算、误差椭球、
// 闭合检测、方向冲突与编号一致性。这是本服务区别于普通 CRUD 的领域核心。
package geometry

import "math"

// Vec3 是三维向量/点，单位由调用方（批次）声明的长度单位决定。
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

// Add 向量加法。
func (v Vec3) Add(o Vec3) Vec3 { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }

// Sub 向量减法。
func (v Vec3) Sub(o Vec3) Vec3 { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }

// Scale 标量乘法。
func (v Vec3) Scale(s float64) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }

// Dot 点积。
func (v Vec3) Dot(o Vec3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }

// Cross 叉积。
func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		v.Y*o.Z - v.Z*o.Y,
		v.Z*o.X - v.X*o.Z,
		v.X*o.Y - v.Y*o.X,
	}
}

// Norm 欧氏范数（长度）。
func (v Vec3) Norm() float64 { return math.Sqrt(v.Dot(v)) }

// Distance 两点间欧氏距离。
func (v Vec3) Distance(o Vec3) float64 { return v.Sub(o).Norm() }

// Unit 单位化；零向量返回零向量，避免除零。
func (v Vec3) Unit() Vec3 {
	n := v.Norm()
	if n < 1e-12 {
		return Vec3{}
	}
	return v.Scale(1 / n)
}

// AngleDeg 两向量夹角（度，0~180）。
func AngleDeg(a, b Vec3) float64 {
	na, nb := a.Norm(), b.Norm()
	if na < 1e-12 || nb < 1e-12 {
		return 0
	}
	cos := a.Dot(b) / (na * nb)
	if cos > 1 {
		cos = 1
	}
	if cos < -1 {
		cos = -1
	}
	return math.Acos(cos) * 180 / math.Pi
}

// Centroid 计算一组点的质心；空集返回零向量。
func Centroid(pts []Vec3) Vec3 {
	if len(pts) == 0 {
		return Vec3{}
	}
	var sum Vec3
	for _, p := range pts {
		sum = sum.Add(p)
	}
	return sum.Scale(1 / float64(len(pts)))
}
