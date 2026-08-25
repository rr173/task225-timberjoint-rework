package geometry

import (
	"sort"
	"strings"
)

// NumberingResult 是编号一致性检查的结果。
type NumberingResult struct {
	Duplicates []string // 重复编号列表
	Malformed  []string // 格式非法编号列表
	OK         bool     // 是否通过（无重复、无格式错误）
}

// CheckNumbering 检查一组编号的一致性：
//  1. 编号不允许为空；
//  2. 编号不允许重复（区分大小写）；
//  3. 编号需符合前缀 + 数字的规范（如 P001、M002、N003）。
//
// prefix 用于报告语义（如 "测点"/"构件"/"节点"），不参与格式判定；
// 允许自定义前缀段，但必须以非空字母/下划线开头。
func CheckNumbering(ids []string) NumberingResult {
	res := NumberingResult{}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			res.Malformed = append(res.Malformed, "(空编号)")
			continue
		}
		if _, ok := seen[trimmed]; ok {
			res.Duplicates = append(res.Duplicates, trimmed)
			continue
		}
		seen[trimmed] = struct{}{}
		if !validNumbering(trimmed) {
			res.Malformed = append(res.Malformed, trimmed)
		}
	}
	sort.Strings(res.Duplicates)
	sort.Strings(res.Malformed)
	res.OK = len(res.Duplicates) == 0 && len(res.Malformed) == 0
	return res
}

// validNumbering 判断编号是否符合「非空标识符 + 数字结尾」的测绘编号规范。
func validNumbering(id string) bool {
	if len(id) < 2 {
		return false
	}
	if !isIdentStart(id[0]) {
		return false
	}
	hasDigit := false
	for i := 1; i < len(id); i++ {
		c := id[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
			continue
		}
		if isIdentChar(c) {
			continue
		}
		return false
	}
	return hasDigit
}

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9') || c == '-' || c == '.'
}
