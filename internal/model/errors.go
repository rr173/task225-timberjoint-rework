// Package model 定义古建筑木构节点测绘复核台的领域实体、
// 状态常量与领域错误。实体为纯数据载体，不含持久化与 HTTP 逻辑，
// 几何计算规则由 geometry 包承载，业务流程由 survey/member/joint/
// evidence/version 各业务包承载。
package model

import "errors"

// 领域错误哨兵：service 与 httpapi 依据它们做错误映射。
var (
	// ErrNotFound 目标资源不存在。
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput 入参非法（坐标缺失、坐标系未声明、时间倒退等）。
	ErrInvalidInput = errors.New("invalid input")
	// ErrInvalidState 状态机不允许的流转（如对已封存批次继续写入）。
	ErrInvalidState = errors.New("invalid state transition")
	// ErrConflict 并发冲突或重复写入（构件编号冲突、节点关系冲突等）。
	ErrConflict = errors.New("conflict")
	// ErrSealed 批次已封存，不可修改。
	ErrSealed = errors.New("sealed")
	// ErrDuplicate 幂等指纹冲突（重复测点/重复证据）。
	ErrDuplicate = errors.New("duplicate")
	// ErrInsufficientData 数据不足（无法闭合校验或发布版本）。
	ErrInsufficientData = errors.New("insufficient data")
	// ErrGeometry 几何校验失败（闭合不满足、方向冲突未解决）。
	ErrGeometry = errors.New("geometry check failed")
)
