package validate

import "regexp"

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValidPhone 校验中国大陆手机号格式
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
