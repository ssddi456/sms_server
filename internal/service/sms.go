package service

import (
	"regexp"
	"strings"
)

func ExtractCode(s string) string {
	content := strings.TrimSpace(s)
	if content == "" {
		return ""
	}

	patterns := []string{
		`(?i)(?:验证码|auth|code|动态密码|短信验证码)[^\d]{0,30}(\d{4,8})`,
		`(?i)(?:verify|otp|token|one[- ]?time[- ]?password)[^\d]{0,30}(\d{4,8})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(content); len(match) > 1 {
			return match[1]
		}
	}

	fallback := regexp.MustCompile(`\d{4,8}`)
	if match := fallback.FindStringSubmatch(content); len(match) > 0 {
		return match[0]
	}

	return ""
}

func ExtractPhoneNumber(simSlot string) string {
	simSlot = strings.TrimSpace(simSlot)
	if simSlot == "" {
		return ""
	}

	phoneRe := regexp.MustCompile(`1[3-9]\d{9}`)
	if match := phoneRe.FindString(simSlot); match != "" {
		return match
	}

	return ""
}
