package scanner

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/response"
)

// Scanner 扫描器接口
type Scanner interface {
	// Setup 初始化扫描器
	Setup(ctx context.Context) error

	// Check 检查响应是否有效
	Check(resp *requester.Response) bool

	// IsWildcard 判断是否为通配符响应
	IsWildcard(resp *requester.Response) bool
}

// BaseScanner 基础扫描器
type BaseScanner struct {
	requester        requester.Requester
	wildcardResp     *requester.Response
	wildcardPath     string
	includeStatusMap map[int]bool
	excludeStatusMap map[int]bool
	excludeSizes     map[string]bool
	excludeTexts     []string
	excludeRegex     *regexp.Regexp
	regexCache       map[string]bool
	mu               sync.RWMutex
}

// NewBaseScanner 创建基础扫描器
func NewBaseScanner(req requester.Requester) *BaseScanner {
	return &BaseScanner{
		requester:        req,
		includeStatusMap: make(map[int]bool),
		excludeStatusMap: make(map[int]bool),
		excludeSizes:     make(map[string]bool),
		regexCache:       make(map[string]bool),
	}
}

// Setup 初始化扫描器
func (s *BaseScanner) Setup(ctx context.Context) error {
	// TODO: 实现通配符检测
	return nil
}

// Check 检查响应是否有效
func (s *BaseScanner) Check(resp *requester.Response) bool {
	// 检查状态码
	if !s.checkStatusCode(resp.Status) {
		return false
	}

	// 检查大小
	if !s.checkSize(resp.Length) {
		return false
	}

	// 检查内容
	if !s.checkContent(resp.Content) {
		return false
	}

	// 检查正则
	if !s.checkRegex(resp.Content) {
		return false
	}

	// 检查重定向
	if !s.checkRedirect(resp.Redirect) {
		return false
	}

	// 检查通配符
	if s.IsWildcard(resp) {
		return false
	}

	return true
}

// IsWildcard 判断是否为通配符响应
func (s *BaseScanner) IsWildcard(resp *requester.Response) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.wildcardResp == nil {
		return false
	}

	// 比较状态码
	if s.wildcardResp.Status != resp.Status {
		return false
	}

	// 比较大小
	if s.wildcardResp.Length != resp.Length {
		return false
	}

	// 简单的内容比较
	// TODO: 实现更复杂的动态内容比较
	return s.wildcardResp.Content == resp.Content
}

// checkStatusCode 检查状态码
func (s *BaseScanner) checkStatusCode(status int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查排除列表
	if s.excludeStatusMap[status] {
		return false
	}

	// 如果有包含列表，检查是否在其中
	if len(s.includeStatusMap) > 0 {
		return s.includeStatusMap[status]
	}

	return true
}

// checkSize 检查大小
func (s *BaseScanner) checkSize(size int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查排除的大小
	sizeStr := response.GetSize(size)
	sizeStr = strings.TrimSuffix(sizeStr, " B")
	sizeStr = strings.TrimSuffix(sizeStr, " KB")
	sizeStr = strings.TrimSuffix(sizeStr, " MB")
	sizeStr = strings.TrimSuffix(sizeStr, " GB")

	if s.excludeSizes[sizeStr] {
		return false
	}

	return true
}

// checkContent 检查内容
func (s *BaseScanner) checkContent(content string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, text := range s.excludeTexts {
		if strings.Contains(content, text) {
			return false
		}
	}

	return true
}

// checkRegex 检查正则表达式
func (s *BaseScanner) checkRegex(content string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.excludeRegex == nil {
		return true
	}

	// 检查缓存
	if cached, ok := s.regexCache[content]; ok {
		return cached
	}

	// 执行正则匹配
	matched := s.excludeRegex.MatchString(content)
	s.regexCache[content] = matched

	return !matched
}

// checkRedirect 检查重定向
func (s *BaseScanner) checkRedirect(redirect string) bool {
	// TODO: 实现重定向过滤
	return true
}

// SetIncludeStatus 设置包含的状态码
func (s *BaseScanner) SetIncludeStatus(statuses []int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.includeStatusMap = make(map[int]bool)
	for _, status := range statuses {
		s.includeStatusMap[status] = true
	}
}

// SetExcludeStatus 设置排除的状态码
func (s *BaseScanner) SetExcludeStatus(statuses []int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.excludeStatusMap = make(map[int]bool)
	for _, status := range statuses {
		s.excludeStatusMap[status] = true
	}
}

// SetExcludeSizes 设置排除的大小
func (s *BaseScanner) SetExcludeSizes(sizes []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.excludeSizes = make(map[string]bool)
	for _, size := range sizes {
		s.excludeSizes[size] = true
	}
}

// SetExcludeTexts 设置排除的文本
func (s *BaseScanner) SetExcludeTexts(texts []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.excludeTexts = texts
}

// SetExcludeRegex 设置排除的正则
func (s *BaseScanner) SetExcludeRegex(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %w", err)
	}

	s.excludeRegex = re
	s.regexCache = make(map[string]bool) // 清空缓存
	return nil
}

// SetWildcard 设置通配符响应
func (s *BaseScanner) SetWildcard(path string, resp *requester.Response) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.wildcardPath = path
	s.wildcardResp = resp
}

// GetWildcard 获取通配符响应
func (s *BaseScanner) GetWildcard() (path string, resp *requester.Response) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.wildcardPath, s.wildcardResp
}

// FilterChain 过滤器链
type FilterChain struct {
	filters []Filter
}

// Filter 过滤器接口
type Filter interface {
	IsExcluded(resp *requester.Response) bool
}

// NewFilterChain 创建过滤器链
func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]Filter, 0),
	}
}

// AddFilter 添加过滤器
func (c *FilterChain) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

// IsExcluded 检查是否被排除
func (c *FilterChain) IsExcluded(resp *requester.Response) bool {
	for _, filter := range c.filters {
		if filter.IsExcluded(resp) {
			return true
		}
	}
	return false
}

// StatusCodeFilter 状态码过滤器
type StatusCodeFilter struct {
	include map[int]bool
	exclude map[int]bool
}

// NewStatusCodeFilter 创建状态码过滤器
func NewStatusCodeFilter(include, exclude []int) *StatusCodeFilter {
	f := &StatusCodeFilter{
		include: make(map[int]bool),
		exclude: make(map[int]bool),
	}

	for _, status := range include {
		f.include[status] = true
	}
	for _, status := range exclude {
		f.exclude[status] = true
	}

	return f
}

// IsExcluded 检查是否被排除
func (f *StatusCodeFilter) IsExcluded(resp *requester.Response) bool {
	// 检查排除列表
	if f.exclude[resp.Status] {
		return true
	}

	// 如果有包含列表，检查是否在其中
	if len(f.include) > 0 {
		return !f.include[resp.Status]
	}

	return false
}

// SizeFilter 大小过滤器
type SizeFilter struct {
	exclude map[int64]bool
	min     int64
	max     int64
}

// NewSizeFilter 创建大小过滤器
func NewSizeFilter(exclude []string, min, max int64) *SizeFilter {
	f := &SizeFilter{
		exclude: make(map[int64]bool),
		min:    min,
		max:    max,
	}

	// 解析大小字符串
	for _, sizeStr := range exclude {
		var size int64
		fmt.Sscanf(sizeStr, "%d", &size)
		f.exclude[size] = true
	}

	return f
}

// IsExcluded 检查是否被排除
func (f *SizeFilter) IsExcluded(resp *requester.Response) bool {
	// 检查排除列表
	if f.exclude[resp.Length] {
		return true
	}

	// 检查最小值
	if f.min > 0 && resp.Length < f.min {
		return true
	}

	// 检查最大值
	if f.max > 0 && resp.Length > f.max {
		return true
	}

	return false
}

// ContentFilter 内容过滤器
type ContentFilter struct {
	excludeTexts []string
}

// NewContentFilter 创建内容过滤器
func NewContentFilter(texts []string) *ContentFilter {
	return &ContentFilter{
		excludeTexts: texts,
	}
}

// IsExcluded 检查是否被排除
func (f *ContentFilter) IsExcluded(resp *requester.Response) bool {
	for _, text := range f.excludeTexts {
		if strings.Contains(resp.Content, text) {
			return true
		}
	}
	return false
}

// RegexFilter 正则过滤器
type RegexFilter struct {
	pattern *regexp.Regexp
}

// NewRegexFilter 创建正则过滤器
func NewRegexFilter(pattern string) (*RegexFilter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &RegexFilter{pattern: re}, nil
}

// IsExcluded 检查是否被排除
func (f *RegexFilter) IsExcluded(resp *requester.Response) bool {
	return f.pattern.MatchString(resp.Content)
}
