package config

// 极简字符串工具，避免引入第三方依赖。

import (
	"strings"
)

// splitLines 按换行符分割字符串。
func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimSuffix(line, "\r"))
	}
	return lines
}

// trimSpace 去除首尾空白。
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// trimQuotes 去除首尾引号。
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// indexOf 返回字符在字符串中的位置，找不到返回 -1。
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// atoi 将字符串转为整数，失败返回默认值。
func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	neg := false
	start := 0
	if s[0] == '-' {
		neg = true
		start = 1
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return def
		}
		n = n*10 + int(s[i]-'0')
	}
	if neg {
		return -n
	}
	return n
}
