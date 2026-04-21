package controller

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/utils"
)

// RecursiveScanner 递归扫描器
type RecursiveScanner struct {
	controller    *Controller
	maxDepth      int
	currentDepth  int
	visitedPaths  map[string]bool
	pathsQueue    []string
	mu            sync.RWMutex
	requester     requester.Requester
	baseURL       string
	foundDirs     map[string]bool
}

// NewRecursiveScanner 创建递归扫描器
func NewRecursiveScanner(ctrl *Controller, maxDepth int) *RecursiveScanner {
	return &RecursiveScanner{
		controller:   ctrl,
		maxDepth:     maxDepth,
		currentDepth: 0,
		visitedPaths: make(map[string]bool),
		pathsQueue:   make([]string, 0),
		requester:    ctrl.requester,
		foundDirs:    make(map[string]bool),
	}
}

// Scan 扫描并发现新目录
func (rs *RecursiveScanner) Scan(ctx context.Context, basePath string) ([]string, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.currentDepth >= rs.maxDepth {
		return nil, fmt.Errorf("达到最大递归深度: %d", rs.maxDepth)
	}

	// 从字典获取所有路径
	dict := rs.controller.dictionary
	var newDirs []string

	for {
		path, ok := dict.Next()
		if !ok {
			break
		}

		// 跳过已访问的路径
		if rs.visitedPaths[path] {
			continue
		}
		rs.visitedPaths[path] = true

		// 发送请求
		resp, err := rs.requester.Request(ctx, path)
		if err != nil {
			continue
		}

		// 检查是否是目录（状态码 200 + 以 / 结尾的路径）
		if rs.isDirectory(resp) {
			rs.foundDirs[path] = true
			newDirs = append(newDirs, path)
		}
	}

	// 递归扫描发现的目录
	if len(newDirs) > 0 && rs.currentDepth < rs.maxDepth {
		rs.currentDepth++

		for _, dir := range newDirs {
			subDirs, err := rs.Scan(ctx, dir)
			if err != nil {
				continue
			}
			newDirs = append(newDirs, subDirs...)
		}

		rs.currentDepth--
	}

	return newDirs, nil
}

// isDirectory 检查响应是否表示目录
func (rs *RecursiveScanner) isDirectory(resp *requester.Response) bool {
	// 方法1: 状态码 200 且路径以 / 结尾
	if resp.Status == 200 && strings.HasSuffix(resp.Path, "/") {
		return true
	}

	// 方法2: Content-Type 包含目录相关
	contentType := strings.ToLower(resp.ContentType)
	if strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") {
		// 检查响应体是否包含目录列表特征
		if rs.containsDirectoryListing(resp.Content) {
			return true
		}
	}

	// 方法3: URL 重定向到类似路径
	if resp.Redirect != "" {
		redirectPath := parsePath(resp.Redirect)
		if strings.HasSuffix(redirectPath, "/") &&
			redirectPath != resp.Path &&
			redirectPath != resp.Path+"/" {
			return true
		}
	}

	return false
}

// containsDirectoryListing 检查响应体是否包含目录列表
func (rs *RecursiveScanner) containsDirectoryListing(content string) bool {
	// 检查常见的目录列表模式
	patterns := []string{
		`<a href="[^"]*/"[^>]*>`,         // HTML 链接以 / 结尾
		`Parent Directory`,               // 父目录链接
		`Directory Listing`,              // 目录列表标题
		`Index of /`,                     // Apache 索引
		`<title>Index of`,                 // HTML 标题
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			// 检查是否有多个匹配（真正的目录列表）
			matches := regexp.MustCompile(pattern).FindAllString(content, -1)
			if len(matches) >= 3 { // 至少3个链接
				return true
			}
		}
	}

	return false
}

// ExtractPaths 从响应中提取路径
func (rs *RecursiveScanner) ExtractPaths(resp *requester.Response) []string {
	var paths []string

	// 从HTML中提取链接
	if strings.Contains(resp.ContentType, "text/html") {
		links := extractHTMLLinks(resp.Content)
		for _, link := range links {
			if path := parsePathFromLink(link); path != "" {
				paths = append(paths, path)
			}
		}
	}

	// 从JSON中提取路径（如果是 API 响应）
	if strings.Contains(resp.ContentType, "application/json") {
		jsonPaths := extractJSONPaths(resp.Content)
		paths = append(paths, jsonPaths...)
	}

	return utils.DeduplicatePaths(paths)
}

// extractHTMLLinks 从HTML中提取链接
func extractHTMLLinks(content string) []string {
	var links []string

	// 简单的链接提取正则
	linkRegex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])
		}
	}

	return links
}

// extractJSONPaths 从JSON中提取路径
func extractJSONPaths(content string) []string {
	var paths []string

	// 简单的JSON路径提取
	// 查找 "url" 或 "path" 字段
	pathRegex := regexp.MustCompile(`"(?:url|path)"\s*:\s*"([^"]+)"`)
	matches := pathRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paths = append(paths, match[1])
		}
	}

	return paths
}

// parsePathFromLink 从链接中解析路径
func parsePathFromLink(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	// 只返回路径部分
	path := u.Path

	// 标准化路径
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return ""
	}

	// 移除查询参数和片段
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	if idx := strings.Index(path, "#"); idx != -1 {
		path = path[:idx]
	}

	return path
}

// parsePath 解析URL路径
func parsePath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Path
}

// GetDiscoveredDirectories 获取发现的目录
func (rs *RecursiveScanner) GetDiscoveredDirectories() []string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	dirs := make([]string, 0, len(rs.foundDirs))
	for dir := range rs.foundDirs {
		dirs = append(dirs, dir)
	}

	return dirs
}

// AddPathsToDictionary 将发现的路径添加到字典
func (rs *RecursiveScanner) AddPathsToDictionary(paths []string) error {
	// TODO: 实现将发现的路径添加到字典的逻辑
	return nil
}

// GetCurrentDepth 获取当前递归深度
func (rs *RecursiveScanner) GetCurrentDepth() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.currentDepth
}

// SetMaxDepth 设置最大递归深度
func (rs *RecursiveScanner) SetMaxDepth(depth int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.maxDepth = depth
}

// IsVisited 检查路径是否已访问
func (rs *RecursiveScanner) IsVisited(path string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.visitedPaths[path]
}

// MarkVisited 标记路径为已访问
func (rs *RecursiveScanner) MarkVisited(path string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.visitedPaths[path] = true
}
