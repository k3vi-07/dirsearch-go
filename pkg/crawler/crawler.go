package crawler

import (
	"regexp"
	"strings"
)

// Crawler 爬虫
type Crawler struct{}

// NewCrawler 创建爬虫
func NewCrawler() *Crawler {
	return &Crawler{}
}

// CrawlResult 爬取结果
type CrawlResult struct {
	Paths []string
}

// Crawl 爬取路径
func (c *Crawler) Crawl(url, scope, content string, crawlType string) []string {
	var results []string

	switch crawlType {
	case "html":
		results = c.htmlCrawl(url, scope, content)
	case "robots":
		results = c.robotsCrawl(content)
	case "text":
		results = c.textCrawl(url, scope, content)
	default:
		results = c.htmlCrawl(url, scope, content)
	}

	return c.filter(results)
}

// htmlCrawl HTML爬取
func (c *Crawler) htmlCrawl(url, scope, content string) []string {
	results := make([]string, 0)

	// 简单的HTML标签解析（不使用外部库）
	// 查找 href="..." 和 src="..."
	hrefRegex := regexp.MustCompile(`href=["']([^"']+)["']`)
	srcRegex := regexp.MustCompile(`src=["']([^"']+)["']`)
	actionRegex := regexp.MustCompile(`action=["']([^"']+)["']`)

	// 提取所有匹配
	results = append(results, c.extractPaths(hrefRegex, url, scope, content)...)
	results = append(results, c.extractPaths(srcRegex, url, scope, content)...)
	results = append(results, c.extractPaths(actionRegex, url, scope, content)...)

	return results
}

// robotsCrawl robots.txt爬取
func (c *Crawler) robotsCrawl(content string) []string {
	results := make([]string, 0)

	// 匹配 Allow: /path 或 Disallow: /path
	regex := regexp.MustCompile(`(?:Allow|Disallow):\s*/([^\s\r\n]+)`)
	matches := regex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			path := "/" + match[1]
			results = append(results, path)
		}
	}

	return results
}

// textCrawl 纯文本爬取
func (c *Crawler) textCrawl(url, scope, content string) []string {
	results := make([]string, 0)

	// 提取 /path 格式的URL
	if !strings.HasSuffix(scope, "/") {
		scope += "/"
	}

	// 匹配 scope + 路径
	regex := regexp.MustCompile(regexp.QuoteMeta(scope) + `[a-zA-Z0-9-._~!$&*+,;=:@?%]+`)
	matches := regex.FindAllString(content, -1)

	for _, match := range matches {
		path := strings.TrimPrefix(match, scope)
		if path != "" && path != "/" {
			results = append(results, path)
		}
	}

	return results
}

// extractPaths 从正则匹配中提取路径
func (c *Crawler) extractPaths(regex *regexp.Regexp, url, scope, content string) []string {
	results := make([]string, 0)

	matches := regex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			path := match[1]

			// 处理相对路径
			if strings.HasPrefix(path, "/") {
				// 绝对路径：/admin -> admin
				path = strings.TrimPrefix(path, "/")
			} else if strings.HasPrefix(path, scope) {
				// 完整URL：http://example.com/admin -> admin
				path = strings.TrimPrefix(path, scope)
			} else if !strings.Contains(path, "://") {
				// 相对路径：admin -> admin
				// 保持不变
			} else {
				// 外部URL，跳过
				continue
			}

			if path != "" && path != "/" {
				results = append(results, path)
			}
		}
	}

	return results
}

// filter 过滤和去重
func (c *Crawler) filter(results []string) []string {
	seen := make(map[string]bool)
	filtered := make([]string, 0)

	for _, result := range results {
		// 去重
		if seen[result] {
			continue
		}
		seen[result] = true

		// 过滤空值和特殊字符
		if result == "" || result == "/" || strings.HasPrefix(result, "#") {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}
