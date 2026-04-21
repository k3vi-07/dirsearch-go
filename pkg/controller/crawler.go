package controller

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/youruser/dirsearch-go/pkg/dictionary"
	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/utils"
)

// Crawler 爬虫
type Crawler struct {
	requester   requester.Requester
	maxDepth    int
	maxPages    int
	visited     map[string]bool
	foundPaths  []string
	mu          sync.RWMutex
	depthMap    map[string]int
}

// NewCrawler 创建爬虫
func NewCrawler(req requester.Requester, maxDepth, maxPages int) *Crawler {
	return &Crawler{
		requester:  req,
		maxDepth:  maxDepth,
		maxPages:   maxPages,
		visited:    make(map[string]bool),
		foundPaths: make([]string, 0),
		depthMap:   make(map[string]int),
	}
}

// Crawl 爬取URL
func (c *Crawler) Crawl(ctx context.Context, startURL string) ([]string, error) {
	return c.crawlRecursive(ctx, startURL, 0)
}

// crawlRecursive 递归爬取
func (c *Crawler) crawlRecursive(ctx context.Context, url string, depth int) ([]string, error) {
	// 检查深度
	if depth > c.maxDepth {
		return nil, nil
	}

	// 检查是否已访问
	c.mu.RLock()
	visited := c.visited[url]
	c.mu.RUnlock()

	if visited {
		return nil, nil
	}

	// 标记为已访问
	c.mu.Lock()
	c.visited[url] = true
	c.depthMap[url] = depth
	c.mu.Unlock()

	// 发送请求
	resp, err := c.requester.Request(ctx, "")
	if err != nil {
		return nil, err
	}

	// 解析响应
	paths := c.extractPaths(resp)

	// 过滤和添加路径
	var newPaths []string
	for _, path := range paths {
		// 跳过已访问的
		if c.isVisited(path) {
			continue
		}

		// 跳过外部链接
		if c.isExternalLink(path) {
			continue
		}

		// 跳过静态资源
		if c.isStaticResource(path) {
			continue
		}

		newPaths = append(newPaths, path)
		c.foundPaths = append(c.foundPaths, path)
	}

	// 递归爬取
	for _, path := range newPaths {
		if len(c.foundPaths) >= c.maxPages {
			break
		}

		subPaths, err := c.crawlRecursive(ctx, path, depth+1)
		if err != nil {
			continue
		}

		c.foundPaths = append(c.foundPaths, subPaths...)
	}

	return c.foundPaths, nil
}

// extractPaths 从响应中提取路径
func (c *Crawler) extractPaths(resp *requester.Response) []string {
	var paths []string

	contentType := strings.ToLower(resp.ContentType)

	// HTML 响应
	if strings.Contains(contentType, "text/html") {
		htmlPaths := c.extractHTMLPaths(resp.Content)
		paths = append(paths, htmlPaths...)
	}

	// JSON 响应
	if strings.Contains(contentType, "application/json") {
		jsonPaths := c.extractJSONPaths(resp.Content)
		paths = append(paths, jsonPaths...)
	}

	// 文本响应
	if strings.Contains(contentType, "text/plain") {
		textPaths := c.extractTextPaths(resp.Content)
		paths = append(paths, textPaths...)
	}

	return utils.DeduplicatePaths(paths)
}

// extractHTMLPaths 从HTML中提取路径
func (c *Crawler) extractHTMLPaths(content string) []string {
	var paths []string

	// 提取所有 href 属性
	hrefRegex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"`)
	matches := hrefRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			href := match[1]
			if path := c.cleanPath(href); path != "" {
				paths = append(paths, path)
			}
		}
	}

	// 提取表单 action 属性
	formRegex := regexp.MustCompile(`<form\s+(?:[^>]*?\s+)?action="([^"]*)"`)
	matches = formRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			action := match[1]
			if path := c.cleanPath(action); path != "" {
				paths = append(paths, path)
			}
		}
	}

	return paths
}

// extractJSONPaths 从JSON中提取路径
func (c *Crawler) extractJSONPaths(content string) []string {
	var paths []string

	// 提取 "url", "path", "href" 字段
	fieldRegex := regexp.MustCompile(`"(?:url|path|href)"\s*:\s*"([^"]+)"`)
	matches := fieldRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			if path := c.cleanPath(match[1]); path != "" {
				paths = append(paths, path)
			}
		}
	}

	return paths
}

// extractTextPaths 从文本中提取路径
func (c *Crawler) extractTextPaths(content string) []string {
	var paths []string

	// 查找 /开头/结尾的路径
	pathRegex := regexp.MustCompile(`(/[a-zA-Z0-9._-]+/?)`)
	matches := pathRegex.FindAllString(content, -1)

	for _, match := range matches {
		if len(match) > 2 { // 至少 /a/
			paths = append(paths, match)
		}
	}

	return paths
}

// cleanPath 清理路径
func (c *Crawler) cleanPath(rawPath string) string {
	// 移除片段
	if idx := strings.Index(rawPath, "#"); idx != -1 {
		rawPath = rawPath[:idx]
	}

	// 移除查询参数（可选）
	if idx := strings.Index(rawPath, "?"); idx != -1 {
		rawPath = rawPath[:idx]
	}

	// 标准化
	rawPath = strings.TrimSuffix(rawPath, "/")

	// 移除协议和域名
	if strings.HasPrefix(rawPath, "http://") {
		rawPath = rawPath[7:]
		if idx := strings.Index(rawPath, "/"); idx != -1 {
			rawPath = rawPath[idx:]
		}
	} else if strings.HasPrefix(rawPath, "https://") {
		rawPath = rawPath[8:]
		if idx := strings.Index(rawPath, "/"); idx != -1 {
			rawPath = rawPath[idx:]
		}
	}

	// 只返回相对路径
	if rawPath == "" || rawPath == "/" {
		return ""
	}

	return rawPath
}

// isVisited 检查是否已访问
func (c *Crawler) isVisited(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.visited[path] || c.visited["/"+path]
}

// isExternalLink 检查是否是外部链接
func (c *Crawler) isExternalLink(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}

// isStaticResource 检查是否是静态资源
func (c *Crawler) isStaticResource(path string) bool {
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico",
		".svg", ".woff", ".woff2", ".ttf", ".eot",
		".mp4", ".mp3", ".avi", ".mov", ".pdf",
		".zip", ".tar", ".gz", ".xml", ".json",
	}

	path = strings.ToLower(path)
	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// GetFoundPaths 获取发现的所有路径
func (c *Crawler) GetFoundPaths() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.foundPaths))
	copy(result, c.foundPaths)
	return result
}

// GetVisitedCount 获取已访问数量
func (c *Crawler) GetVisitedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.visited)
}

// GetDepthMap 获取深度映射
func (c *Crawler) GetDepthMap() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int, len(c.depthMap))
	for k, v := range c.depthMap {
		result[k] = v
	}
	return result
}

// MergeIntoDictionary 合并到字典
func (c *Crawler) MergeIntoDictionary(dict *dictionary.Dictionary, maxEntries int) error {
	paths := c.GetFoundPaths()

	// 限制数量
	if maxEntries > 0 && len(paths) > maxEntries {
		paths = paths[:maxEntries]
	}

	// TODO: 添加到字典
	// dict.AddPaths(paths)

	return nil
}

// ExportToFiles 导出到文件
func (c *Crawler) ExportToFiles(filepath string) error {
	paths := c.GetFoundPaths()

	// 写入文件
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, path := range paths {
		fmt.Fprintln(file, path)
	}

	return nil
}

// GenerateReport 生成爬取报告
func (c *Crawler) GenerateReport() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"total_found":  len(c.foundPaths),
		"visited":     len(c.visited),
		"max_depth":    c.maxDepth,
		"depth_map":    c.depthMap,
	}
}
