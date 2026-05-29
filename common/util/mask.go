package util

import "strings"

// MaskPhone 手机号脱敏：13812345678 → 138****5678
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return "***"
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskName 姓名脱敏：张三 → 张*，欧阳三 → 欧阳*
// 保留第一个字，其余用 * 替代
func MaskName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}
