// Package joint 承载节点关系（木构节点连接）与几何复核的业务规则。
package joint

import (
	"encoding/json"
	"fmt"
	"strings"

	"task225-timberjoint/internal/model"
)

// 节点关系状态机允许的流转表。
var jointTransitions = map[string][]string{
	model.JointCandidate: {model.JointClosed, model.JointBroken, model.JointRejected},
	model.JointClosed:    {model.JointConfirmed, model.JointRejected},
	model.JointBroken:    {model.JointConfirmed, model.JointRejected},
	model.JointConfirmed: {},
	model.JointRejected:  {},
}

// CanTransition 判断节点关系状态机是否允许 from -> to。
func CanTransition(from, to string) bool {
	for _, t := range jointTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// CanCheck reports whether a relation is still editable by a fresh review.
func CanCheck(status string) bool {
	return status == model.JointCandidate || status == model.JointClosed || status == model.JointBroken
}

// ParseIDList 解析 JSON 数组字符串为 ID 列表；空串返回空列表。
func ParseIDList(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, fmt.Errorf("%w: invalid id list: %v", model.ErrInvalidInput, err)
	}
	return ids, nil
}

// EncodeIDList 序列化 ID 列表为 JSON 数组字符串。
func EncodeIDList(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(ids)
	return string(b)
}
