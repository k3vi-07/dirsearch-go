package scanner

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/utils"
)

const (
	// WILDCARD_TEST_POINT_MARKER 通配符测试点标记
	WILDCARD_TEST_POINT_MARKER = "__WILDCARD_POINT__"
	// REFLECTED_PATH_MARKER 反射路径标记
	REFLECTED_PATH_MARKER = "__REFLECTED_PATH__"
	// TEST_PATH_LENGTH 测试路径长度
	TEST_PATH_LENGTH = 16
)

// WildcardScanner 通配符扫描器
type WildcardScanner struct {
	requester       requester.Requester
	basePath        string
	wildcardResp    *requester.Response
	contentParser   *ContentParser
	redirectRegex   string
	mu              sync.RWMutex
}

// NewWildcardScanner 创建通配符扫描器
func NewWildcardScanner(req requester.Requester, basePath string) *WildcardScanner {
	return &WildcardScanner{
		requester: req,
		basePath:  basePath,
	}
}

// Setup 初始化通配符检测
func (s *WildcardScanner) Setup(ctx context.Context) error {
	// 生成两个随机测试路径
	path1 := s.generateRandomPath()
	path2 := s.generateRandomPath()

	// 发送第一个测试请求
	resp1, err := s.requester.Request(ctx, path1)
	if err != nil {
		return fmt.Errorf("通配符测试请求失败: %w", err)
	}

	// 发送第二个测试请求
	resp2, err := s.requester.Request(ctx, path2)
	if err != nil {
		return fmt.Errorf("通配符测试请求失败: %w", err)
	}

	// 检查是否是通配符
	if s.isWildcardResponse(resp1, resp2) {
		s.mu.Lock()
		s.wildcardResp = resp1
		s.mu.Unlock()

		// 创建内容解析器
		s.contentParser = NewContentParser(resp1.Content, resp2.Content)

		// 生成重定向正则
		if resp1.Redirect != "" && resp2.Redirect != "" {
			s.redirectRegex = s.generateRedirectRegex(
				resp1.Redirect, path1,
				resp2.Redirect, path2,
			)
		}

		return nil
	}

	return nil
}

// Check 检查响应是否为通配符
func (s *WildcardScanner) Check(resp *requester.Response) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.wildcardResp == nil {
		return true // 没有通配符，所有响应都有效
	}

	// 检查状态码
	if s.wildcardResp.Status != resp.Status {
		return true
	}

	// 检查重定向
	if s.redirectRegex != "" && resp.Redirect != "" {
		matched, _ := regexp.MatchString(s.redirectRegex, resp.Redirect)
		if !matched {
			return true
		}
	}

	// 检查内容
	if s.contentParser != nil {
		return !s.contentParser.IsMatch(resp.Content)
	}

	// 默认比较大小
	return s.wildcardResp.Length != resp.Length
}

// IsWildcard 是否检测到通配符
func (s *WildcardScanner) IsWildcard() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wildcardResp != nil
}

// GetWildcardResp 获取通配符响应
func (s *WildcardScanner) GetWildcardResp() *requester.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wildcardResp
}

// isWildcardResponse 检查是否是通配符响应
func (s *WildcardScanner) isWildcardResponse(resp1, resp2 *requester.Response) bool {
	// 相同状态码
	if resp1.Status != resp2.Status {
		return false
	}

	// 相同大小
	if resp1.Length != resp2.Length {
		return false
	}

	// 相同内容
	if resp1.Content == resp2.Content {
		return true
	}

	return false
}

// generateRandomPath 生成随机测试路径
func (s *WildcardScanner) generateRandomPath() string {
	randomPart := utils.RandomString(TEST_PATH_LENGTH)
	return s.basePath + WILDCARD_TEST_POINT_MARKER + randomPart
}

// generateRedirectRegex 生成重定向正则表达式
func (s *WildcardScanner) generateRedirectRegex(loc1, path1, loc2, path2 string) string {
	// 替换路径为标记
	loc1 = s.replacePath(loc1, path1, REFLECTED_PATH_MARKER)
	loc2 = s.replacePath(loc2, path2, REFLECTED_PATH_MARKER)

	return utils.GenerateMatchingRegex(loc1, loc2)
}

// replacePath 替换路径
func (s *WildcardScanner) replacePath(loc, path, replaceWith string) string {
	cleanPath := strings.TrimPrefix(path, "/")
	cleanLoc := strings.TrimPrefix(loc, "/")

	// 尝试完整路径匹配
	if idx := strings.Index(cleanLoc, cleanPath); idx != -1 {
		result := cleanLoc[:idx] + replaceWith + cleanLoc[idx+len(cleanPath):]
		return "/" + result
	}

	return loc
}

// ContentParser 动态内容解析器
type ContentParser struct {
	staticPatterns []string
	isStatic       bool
	baseContent    string
}

// NewContentParser 创建内容解析器
func NewContentParser(content1, content2 string) *ContentParser {
	if content1 == content2 {
		return &ContentParser{
			isStatic:    true,
			baseContent: content1,
		}
	}

	patterns := extractStaticPatterns(content1, content2)

	return &ContentParser{
		staticPatterns: patterns,
		isStatic:       false,
		baseContent:    content1,
	}
}

// IsMatch 检查内容是否匹配
func (p *ContentParser) IsMatch(content string) bool {
	if p.isStatic {
		return content == p.baseContent
	}

	// 检查静态模式
	misses := 0
	maxMisses := 1

	if len(p.staticPatterns) < 20 {
		maxMisses = 0
	}

	for _, pattern := range p.staticPatterns {
		if !strings.Contains(content, pattern) {
			misses++
			if misses > maxMisses {
				return false
			}
		}
	}

	return true
}

// extractStaticPatterns 提取静态模式
func extractStaticPatterns(content1, content2 string) []string {
	// 简单分词比较
	words1 := strings.Fields(content1)
	words2 := strings.Fields(content2)

	patterns := make([]string, 0)

	minLen := len(words1)
	if len(words2) < minLen {
		minLen = len(words2)
	}

	for i := 0; i < minLen; i++ {
		if words1[i] == words2[i] && len(words1[i]) > 3 {
			patterns = append(patterns, words1[i])
		}
	}

	return patterns
}

// generateMatchingRegex 生成匹配正则（工具函数）
func generateMatchingRegex(string1, string2 string) string {
	start := "^"
	end := "$"

	// 前向匹配
	runes1 := []rune(string1)
	runes2 := []rune(string2)

	minLen := len(runes1)
	if len(runes2) < minLen {
		minLen = len(runes2)
	}

	for i := 0; i < minLen; i++ {
		if runes1[i] != runes2[i] {
			start += ".*"
			break
		}
		start += regexp.QuoteMeta(string(runes1[i]))
	}

	// 后向匹配
	if strings.Contains(start, ".*") {
		for i := 1; i <= minLen; i++ {
			if runes1[len(runes1)-i] != runes2[len(runes2)-i] {
				break
			}
			end = regexp.QuoteMeta(string(runes1[len(runes1)-i])) + end
		}
	}

	return start + end
}

// ScannerSet 扫描器集合
type ScannerSet struct {
	scalers map[string]*WildcardScanner
	mu      sync.RWMutex
}

// NewScannerSet 创建扫描器集合
func NewScannerSet() *ScannerSet {
	return &ScannerSet{
		scalers: make(map[string]*WildcardScanner),
	}
}

// Get 获取扫描器
func (s *ScannerSet) Get(key string) (*WildcardScanner, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scanner, ok := s.scalers[key]
	return scanner, ok
}

// Set 设置扫描器
func (s *ScannerSet) Set(key string, scanner *WildcardScanner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scalers[key] = scanner
}

// Range 遍历所有扫描器
func (s *ScannerSet) Range(fn func(key string, scanner *WildcardScanner) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for key, scanner := range s.scalers {
		if !fn(key, scanner) {
			break
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
