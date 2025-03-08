package util

import "regexp"

func MatchIP(ipStr string) (ip string) {
	re := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:/\d{1,2})?`)
	ip = re.FindString(ipStr)
	return
}

func GetRemark(remarkStr string) (remark string) {
	re := regexp.MustCompile(`备注\s+(.*)`)
	match := re.FindStringSubmatch(remarkStr)
	if len(match) > 1 {
		remark = match[1]
	}
	return
}
