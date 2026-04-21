package utils

import (
	"math/rand"
	"strings"
	"time"
)

const (
	// 字符集
	letters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits        = "0123456789"
	alphanumeric  = letters + digits
)

// RandomString 生成随机字符串
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return string(b)
}

// RandomStringExclude 生成随机字符串（排除某些字符）
func RandomStringExclude(n int, exclude string) string {
	charset := excludeChars(alphanumeric, exclude)
	if len(charset) == 0 {
		charset = letters // 回退到字母
	}

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// excludeChars 排除字符
func excludeChars(s, exclude string) string {
	result := make([]byte, 0, len(s))
	excludeMap := make(map[rune]bool)

	for _, c := range exclude {
		excludeMap[c] = true
	}

	for _, c := range s {
		if !excludeMap[c] {
			result = append(result, byte(c))
		}
	}

	return string(result)
}

// GenerateMatchingRegex 生成匹配正则
func GenerateMatchingRegex(string1, string2 string) string {
	start := "^"
	end := "$"

	runes1 := []rune(string1)
	runes2 := []rune(string2)

	minLen := len(runes1)
	if len(runes2) < minLen {
		minLen = len(runes2)
	}

	// 前向匹配
	for i := 0; i < minLen; i++ {
		if runes1[i] != runes2[i] {
			start += ".*"
			break
		}
		if i < minLen-1 {
			start += string(runes1[i])
		}
	}

	// 后向匹配
	if strings.Contains(start, ".*") {
		for i := 1; i <= minLen; i++ {
			if runes1[len(runes1)-i] != runes2[len(runes2)-i] {
				break
			}
			end = string(runes1[len(runes1)-i]) + end
		}
	}

	return start + end
}

// DeduplicatePaths 去重路径切片
func DeduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(paths))

	for _, path := range paths {
		if !seen[path] && path != "" && path != "/" {
			seen[path] = true
			result = append(result, path)
		}
	}

	return result
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
