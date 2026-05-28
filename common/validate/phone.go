package validate

import "regexp"

// PhoneRegex 中国大陆手机号正则（预编译，包级变量，各服务统一引用）
var PhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValidPhone 校验中国大陆手机号格式
func IsValidPhone(phone string) bool {
	return PhoneRegex.MatchString(phone)
}
