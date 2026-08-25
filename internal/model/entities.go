package model

import "time"

// 测绘批次状态机：
// collecting(采集中) -> organizing(整理中) -> reviewing(待复核) -> published(已发布) -> archived(封存)。
// 封存为单向终态，封存后不可再写入测点、构件、关系或修改版本。
const (
	BatchCollecting = "collecting" // 采集中：已声明坐标系，正在接收测点与构件
	BatchOrganizing = "organizing" // 整理中：采集结束，正在整理测点与编号
	BatchReviewing  = "reviewing"  // 待复核：正在做几何闭合与方向复核
	BatchPublished  = "published"  // 已发布：节点版本已发布，供修缮方案引用
	BatchArchived   = "archived"   // 封存：冻结为不可变记录
)

// 构件状态机：
// unmatched(待匹配) -> matched(已匹配) / direction_conflict(方向冲突) / missing(缺失)。
const (
	MemberUnmatched         = "unmatched"          // 待匹配：已登记端点，尚未关联到节点
	MemberMatched           = "matched"            // 已匹配：端点已关联到节点关系
	MemberDirectionConflict = "direction_conflict" // 方向冲突：端点方向与节点主方向夹角超容差
	MemberMissing           = "missing"            // 缺失：构件缺失或端点残缺，不参与闭合
)

// 节点关系状态机：
// candidate(候选) -> closed(闭合) / broken(断裂) -> confirmed(确认) / rejected(否决)。
const (
	JointCandidate = "candidate" // 候选：已建立连接，尚未完成几何校验
	JointClosed    = "closed"    // 闭合：测点与端点空间散布在容差内
	JointBroken    = "broken"    // 断裂：闭合残差超容差，存在结构缺陷
	JointConfirmed = "confirmed" // 确认：复核人员确认闭合结论
	JointRejected  = "rejected"  // 否决：复核人员否决该连接关系
)

// 节点版本状态机：
// draft(草稿) -> shared(共享) -> frozen(冻结) -> superseded(替代)。
const (
	VersionDraft      = "draft"      // 草稿：正在编辑节点快照
	VersionShared     = "shared"     // 共享：已共享给复核人员
	VersionFrozen     = "frozen"     // 冻结：已发布为不可变现场记录
	VersionSuperseded = "superseded" // 替代：被更新版本取代
)

// 测点状态机：
// pending(待校验) -> valid(有效) / contact_error(接触异常) / gap(缺口)。
const (
	PointPending      = "pending"       // 待校验：已入库，未完成坐标与误差校验
	PointValid        = "valid"         // 有效：坐标与误差椭球通过校验
	PointContactError = "contact_error" // 接触异常：测点接触不良，保留原始数据但不参与闭合
	PointGap          = "gap"           // 缺口：数据缺口，不参与闭合判定
)

// SurveyBatch 是一次古建木构节点测绘批次，承载坐标系声明与整体生命周期。
type SurveyBatch struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CoordinateSystem string    `json:"coordinate_system"` // 坐标系声明，如 CGCS2000 / 本地工程坐标
	LengthUnit       string    `json:"length_unit"`       // 长度单位：m / mm
	ClosureTolerance float64   `json:"closure_tolerance"` // 闭合容差（单位与 LengthUnit 一致）
	DirectionTolDeg  float64   `json:"direction_tol_deg"` // 方向冲突判定角度容差（度）
	Status           string    `json:"status"`
	Fingerprint      string    `json:"fingerprint"` // 幂等指纹：名称+坐标系+单位哈希
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SurveyPoint 是一个测点：三维坐标 + 误差椭球三半轴 + 主轴方位角。
type SurveyPoint struct {
	ID               string    `json:"id"`
	BatchID          string    `json:"batch_id"`
	PointNo          string    `json:"point_no"`           // 测点编号，如 P001
	X                float64   `json:"x"`                  // 三维坐标（单位由批次声明）
	Y                float64   `json:"y"`
	Z                float64   `json:"z"`
	ErrorSemiMajor   float64   `json:"error_semi_major"`   // 误差椭球长半轴
	ErrorSemiMinor   float64   `json:"error_semi_minor"`   // 误差椭球短半轴
	ErrorVertical    float64   `json:"error_vertical"`     // 误差椭球垂向半轴
	OrientationAzim  float64   `json:"orientation_azim"`   // 椭球主轴方位角（度）
	Status           string    `json:"status"`
	Fingerprint      string    `json:"fingerprint"` // 幂等指纹：批次+编号+坐标哈希
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Member 是木构构件（柱/梁/枋/斗/拱），用两端点表达其空间位置与方向。
type Member struct {
	ID          string    `json:"id"`
	BatchID     string    `json:"batch_id"`
	MemberNo    string    `json:"member_no"`     // 构件编号，如 M001
	MemberType  string    `json:"member_type"`   // 构件类型：柱/梁/枋/斗/拱/其他
	StartX      float64   `json:"start_x"`       // 起点坐标
	StartY      float64   `json:"start_y"`
	StartZ      float64   `json:"start_z"`
	EndX        float64   `json:"end_x"`         // 终点坐标
	EndY        float64   `json:"end_y"`
	EndZ        float64   `json:"end_z"`
	TenonDesc   string    `json:"tenon_desc"`    // 榫卯描述（JSON：榫头/卯口类型与尺寸）
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JointRelation 是一个节点关系：把一组测点与构件端点关联，并记录闭合/方向复核结论。
type JointRelation struct {
	ID              string    `json:"id"`
	BatchID         string    `json:"batch_id"`
	NodeNo          string    `json:"node_no"`           // 节点编号，如 N001
	PointIDs        string    `json:"point_ids"`         // 关联测点 ID（JSON 数组）
	MemberIDs       string    `json:"member_ids"`        // 关联构件 ID（JSON 数组）
	Status          string    `json:"status"`
	ClosureResidual float64   `json:"closure_residual"`  // 闭合残差：点到质心最大距离
	DirectionSpread float64   `json:"direction_spread"`  // 方向散布：构件方向最大夹角（度）
	CheckReport     string    `json:"check_report"`      // 校验报告（JSON）
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PhotoEvidence 是现场照片证据：记录文件名、说明与内容哈希（幂等）。
type PhotoEvidence struct {
	ID       string    `json:"id"`
	JointID  string    `json:"joint_id"`
	BatchID  string    `json:"batch_id"`
	Filename string    `json:"filename"` // 文件名/路径描述
	Caption  string    `json:"caption"`  // 说明
	Hash     string    `json:"hash"`     // 内容哈希（幂等指纹）
	TakenAt  time.Time `json:"taken_at"` // 现场拍摄时间
	CreatedAt time.Time `json:"created_at"`
}

// NodeVersion 是节点复核版本：冻结测点/构件/关系/证据的不可变快照。
type NodeVersion struct {
	ID           string     `json:"id"`
	BatchID      string     `json:"batch_id"`
	NodeNo       string     `json:"node_no"`        // 节点编号
	VersionNo    int        `json:"version_no"`     // 版本号（同节点递增）
	Status       string     `json:"status"`
	SnapshotJSON string     `json:"snapshot_json"`  // 节点快照（JSON：测点+构件+关系+证据）
	Reason       string     `json:"reason"`         // 发布/冻结原因
	CreatedAt    time.Time  `json:"created_at"`
	PublishedAt  *time.Time `json:"published_at"`   // 冻结时间（可空）
}

// StatSummary 是自检/统计接口返回的全局快照。
type StatSummary struct {
	Batches   int `json:"batches"`
	Points    int `json:"points"`
	Members   int `json:"members"`
	Joints    int `json:"joints"`
	Evidences int `json:"evidences"`
	Versions  int `json:"versions"`
	OpenBatches int `json:"open_batches"`
}
