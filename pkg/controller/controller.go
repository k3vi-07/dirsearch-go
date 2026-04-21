package controller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/youruser/dirsearch-go/pkg/config"
	"github.com/youruser/dirsearch-go/pkg/dictionary"
	"github.com/youruser/dirsearch-go/pkg/fuzzer"
	"github.com/youruser/dirsearch-go/pkg/report"
	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/scanner"
	"github.com/youruser/dirsearch-go/internal/logger"
)

// Controller 主控制器
type Controller struct {
	config         *config.Config
	requester      requester.Requester
	dictionary     *dictionary.Dictionary
	fuzzer         fuzzer.Fuzzer
	scannerSet     *scanner.ScannerSet
	filterChain    *fuzzer.FilterChain
	reporters      []report.Reporter
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
	// 递归扫描相关
	directories    []string      // 待扫描的递归目录
	passedUrls     map[string]bool // 已扫描的URL（去重）
	basePath       string         // 基础路径
	currentDepth   int            // 当前递归深度
	// 会话相关
	results        []*fuzzer.ScanResult // 收集的扫描结果
	startTime      time.Time            // 开始时间
	currentURL     string               // 当前扫描的URL
}

// NewController 创建控制器
func NewController(cfg *config.Config) (*Controller, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	ctrl := &Controller{
		config:       cfg,
		ctx:          ctx,
		cancel:       cancel,
		scannerSet:   scanner.NewScannerSet(),
		directories:  make([]string, 0),
		passedUrls:   make(map[string]bool),
		currentDepth: 0,
		results:      make([]*fuzzer.ScanResult, 0),
		startTime:    time.Now(),
	}

	// 初始化请求器
	if err := ctrl.initRequester(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化请求器失败: %w", err)
	}

	// 初始化字典
	if err := ctrl.initDictionary(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化字典失败: %w", err)
	}

	// 初始化过滤器
	ctrl.initFilters()

	// 初始化扫描器集合
	ctrl.scannerSet = scanner.NewScannerSet()

	// 初始化模糊测试器
	ctrl.initFuzzer()

	// 初始化报告器
	if err := ctrl.initReporters(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化报告器失败: %w", err)
	}

	return ctrl, nil
}

// initRequester 初始化请求器
func (c *Controller) initRequester() error {
	reqCfg := &requester.Config{
		ThreadCount:      c.config.ThreadCount,
		Timeout:          c.config.GetTimeoutDuration(),
		MaxRate:          c.config.MaxRate,
		UserAgent:        c.config.UserAgent,
		Headers:          c.config.Headers,
		Proxy:            c.config.Proxy,
		ProxyAuth:        c.config.ProxyAuth,
		VerifySSL:        true, // TODO: 从配置读取
		FollowRedirects:  false,
		MaxRedirects:     3,
		MaxRetries:       c.config.MaxRetries,
		NetworkInterface: c.config.NetworkIf,
	}

	req, err := requester.NewSyncRequester(reqCfg)
	if err != nil {
		return err
	}

	c.requester = req
	return nil
}

// initDictionary 初始化字典
func (c *Controller) initDictionary() error {
	var dict *dictionary.Dictionary
	var err error

	// 检查是否使用内置字典
	if len(c.config.Wordlists) == 1 && c.config.Wordlists[0] == "__builtin__" {
		// 使用内置默认字典
		dict, err = dictionary.NewWithDefault(
			c.config.Extensions,
			c.config.ForceExtensions,
		)
		if err != nil {
			return fmt.Errorf("加载内置字典失败: %w", err)
		}
	} else {
		// 从文件加载字典
		dict, err = dictionary.NewDictionary(
			c.config.Wordlists,
			c.config.Extensions,
			c.config.ForceExtensions,
		)
		if err != nil {
			return err
		}
	}

	// 设置前缀和后缀
	dict.SetPrefixes(c.config.Prefixes)
	dict.SetSuffixes(c.config.Suffixes)

	// 设置大小写转换
	dict.SetLowercase(c.config.Lowercase)
	dict.SetUppercase(c.config.Uppercase)
	dict.SetCapitalization(c.config.Capitalization)

	c.dictionary = dict
	logger.Infof("字典加载完成: %d 个条目", dict.Length())

	return nil
}

// initFilters 初始化过滤器
func (c *Controller) initFilters() {
	c.filterChain = fuzzer.NewFilterChain()

	// 状态码过滤器
	includeStatus, _ := config.ParseStatusCodes(c.config.IncludeStatusCodes)
	excludeStatus, _ := config.ParseStatusCodes(c.config.ExcludeStatusCodes)
	statusFilter := scanner.NewStatusCodeFilter(includeStatus, excludeStatus)
	c.filterChain.AddFilter(statusFilter)

	// 大小过滤器
	var minSize, maxSize int64
	minSize = c.config.MinResponseSize
	maxSize = c.config.MaxResponseSize

	sizeFilter := scanner.NewSizeFilter(c.config.ExcludeSizes, minSize, maxSize)
	c.filterChain.AddFilter(sizeFilter)

	// 内容过滤器
	if len(c.config.ExcludeTexts) > 0 {
		contentFilter := scanner.NewContentFilter(c.config.ExcludeTexts)
		c.filterChain.AddFilter(contentFilter)
	}

	// 正则过滤器
	if c.config.ExcludeRegex != "" {
		regexFilter, err := scanner.NewRegexFilter(c.config.ExcludeRegex)
		if err == nil {
			c.filterChain.AddFilter(regexFilter)
		} else {
			logger.Warnf("无效的正则表达式: %v", err)
		}
	}
}

// initFuzzer 初始化模糊测试器
func (c *Controller) initFuzzer() {
	// TODO: 实现完整的模糊测试器初始化
	// 这里需要创建实际的 AsyncFuzzer 实例
}

// initReporters 初始化报告器
func (c *Controller) initReporters() error {
	factory := &report.ReporterFactory{}

	// 如果没有指定输出文件，使用控制台plain输出
	if c.config.OutputFile == "" {
		rep, err := factory.NewReporter(report.FormatPlain, "")
		if err != nil {
			return err
		}
		c.reporters = append(c.reporters, rep)
		return nil
	}

	// 为每种输出格式创建报告器
	for _, format := range c.config.OutputFormats {
		var reportFormat report.ReportFormat
		switch format {
		case "json":
			reportFormat = report.FormatJSON
		case "csv":
			reportFormat = report.FormatCSV
		case "xml":
			reportFormat = report.FormatXML
		case "html":
			reportFormat = report.FormatHTML
		case "markdown":
			reportFormat = report.FormatMarkdown
		default:
			reportFormat = report.FormatPlain
		}

		rep, err := factory.NewReporter(reportFormat, c.config.OutputFile)
		if err != nil {
			return err
		}
		c.reporters = append(c.reporters, rep)
	}

	return nil
}

// Run 运行扫描
func (c *Controller) Run() error {
	logger.Info("开始扫描...")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动信号处理goroutine
	go c.handleSignals(sigChan)

	// 为每个URL创建扫描器
	for _, url := range c.config.URLs {
		logger.Infof("扫描目标: %s", url)

		// 设置请求器URL
		c.requester.SetURL(url)

		// 初始化扫描器
		wildcardScanner := scanner.NewWildcardScanner(c.requester, "")
		if err := wildcardScanner.Setup(c.ctx); err != nil {
			logger.Warnf("通配符检测失败: %v", err)
		} else {
			c.scannerSet.Set(url, wildcardScanner)
			if wildcardScanner.IsWildcard() {
				logger.Infof("检测到通配符响应: %d", wildcardScanner.GetWildcardResp().Status)
			}
		}

		// 执行扫描
		if err := c.scanURL(url); err != nil {
			logger.Errorf("扫描失败: %v", err)
		}

		// 检查是否需要退出
		select {
		case <-sigChan:
			logger.Info("收到中断信号，正在停止...")
			c.Stop()
			return nil
		case <-c.ctx.Done():
			return nil
		default:
		}
	}

	logger.Info("扫描完成")
	return nil
}

// scanURL 扫描单个URL
func (c *Controller) scanURL(url string) error {
	c.currentURL = url
	startTime := time.Now()

	// 创建结果通道
	resultChan := make(chan *fuzzer.ScanResult, 100)
	errorChan := make(chan error, 10)

	// 创建worker pool
	var wg sync.WaitGroup
	workers := c.config.ThreadCount

	// 启动结果处理器
	done := make(chan struct{})
	go func() {
		for {
			select {
			case result := <-resultChan:
				c.handleResult(result)
			case err := <-errorChan:
				logger.Errorf("错误: %v", err)
			case <-done:
				return
			}
		}
	}()

	// 启动workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go c.worker(url, resultChan, errorChan, &wg)
	}

	// 等待完成
	wg.Wait()
	close(done)
	close(resultChan)
	close(errorChan)

	duration := time.Since(startTime)
	current, total := c.dictionary.Progress()
	logger.Infof("扫描完成: %d/%d (%.2f%%), 耗时: %v", current, total, float64(current)/float64(total)*100, duration)

	// 关闭所有报告器
	logger.Infof("正在保存报告，共 %d 个报告器...", len(c.reporters))
	for i, reporter := range c.reporters {
		logger.Infof("开始关闭报告器[%d]...", i)
		if err := reporter.Close(); err != nil {
			logger.Errorf("关闭报告器[%d]失败: %v", i, err)
		} else {
			logger.Infof("报告器[%d]保存成功", i)
		}
		logger.Infof("报告器[%d]处理完成", i)
	}
	logger.Infof("所有报告器已关闭")

	return nil
}

// worker 工作协程
func (c *Controller) worker(baseURL string, resultChan chan<- *fuzzer.ScanResult, errorChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// 获取下一个路径
		path, ok := c.dictionary.Next()
		if !ok {
			return
		}

		// 发送请求
		resp, err := c.requester.Request(c.ctx, path)
		if err != nil {
			errorChan <- err
			continue
		}

		// 使用过滤器链检查是否应该包含此结果
		if c.filterChain.IsExcluded(resp) {
			continue
		}

		// 发送结果
		result := &fuzzer.ScanResult{
			URL:         resp.URL,
			Path:        resp.Path,
			Status:      resp.Status,
			Size:        resp.Length,
			ContentType: resp.ContentType,
			Redirect:    resp.Redirect,
			Time:        time.Now(),
		}

		select {
		case resultChan <- result:
		default:
			// 通道满，丢弃
		}
	}
}

// handleResult 处理扫描结果
func (c *Controller) handleResult(result *fuzzer.ScanResult) {
	// 检查result是否为nil
	if result == nil {
		logger.Errorf("收到nil结果，跳过")
		return
	}

	// 输出到日志
	logger.Infof("[+] %d - %s - %d bytes", result.Status, result.Path, result.Size)

	// 收集结果到列表
	c.mu.Lock()
	c.results = append(c.results, result)
	c.mu.Unlock()

	// 发送到报告器
	for _, reporter := range c.reporters {
		if err := reporter.Add(result); err != nil {
			logger.Errorf("报告器写入失败: %v", err)
		}
	}

	// 爬虫检查（如果启用）
	if c.config.Crawl {
		c.checkCrawl(result)
	}

	// 递归扫描检查
	c.checkRecursive(result)
}

// Pause 暂停扫描
func (c *Controller) Pause() error {
	logger.Info("暂停扫描...")
	// TODO: 实现暂停逻辑
	return nil
}

// Resume 恢复扫描
func (c *Controller) Resume() error {
	logger.Info("恢复扫描...")
	// TODO: 实现恢复逻辑
	return nil
}

// Stop 停止扫描
func (c *Controller) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	logger.Info("扫描已停止")
}

// GetProgress 获取进度
func (c *Controller) GetProgress() (current, total int) {
	return c.dictionary.Progress()
}

// GetStats 获取统计信息
func (c *Controller) GetStats() map[string]interface{} {
	current, total := c.dictionary.Progress()

	return map[string]interface{}{
		"total":       total,
		"scanned":     current,
		"progress":    float64(current) / float64(total) * 100,
		"urls_count":  len(c.config.URLs),
		"threads":     c.config.ThreadCount,
	}
}

// checkRecursive 检查是否需要递归扫描
func (c *Controller) checkRecursive(result *fuzzer.ScanResult) {
	// 检查是否启用了递归
	if !c.config.Recursive && !c.config.DeepRecursive && !c.config.ForceRecursive {
		return
	}

	// 检查递归深度
	if c.currentDepth >= c.config.RecursiveDepth {
		return
	}

	// 检查状态码是否在递归状态码列表中
	if !c.isRecursionStatus(result.Status) {
		return
	}

	// 处理递归逻辑
	path := result.Path

	// 强制递归：非目录也当作目录处理
	if c.config.ForceRecursive && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// 根据递归模式添加目录
	var newDirs []string

	if c.config.DeepRecursive {
		// 深度递归：提取路径中所有层级的目录
		newDirs = c.extractAllDirectories(path)
	} else if c.config.Recursive || c.config.ForceRecursive {
		// 标准递归/强制递归：只处理以/结尾的目录
		if strings.HasSuffix(path, "/") {
			newDirs = []string{path}
		}
	}

	// 添加新目录到待扫描列表
	for _, dir := range newDirs {
		c.addDirectory(dir)
	}
}

// isRecursionStatus 检查状态码是否触发递归
func (c *Controller) isRecursionStatus(status int) bool {
	for _, codeStr := range c.config.RecursionStatusCodes {
		var code int
		fmt.Sscanf(codeStr, "%d", &code)
		if status == code {
			return true
		}
	}
	return false
}

// extractAllDirectories 提取路径中所有层级的目录
func (c *Controller) extractAllDirectories(path string) []string {
	dirs := make([]string, 0)

	// 清理路径
	cleanPath := path
	if strings.HasPrefix(cleanPath, "/") {
		cleanPath = cleanPath[1:]
	}

	// 分割路径
	parts := strings.Split(cleanPath, "/")

	// 构建所有层级的目录
	currentPath := ""
	for i, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath += "/" + part
		}

		// 添加当前层级的目录（以/结尾）
		if i < len(parts)-1 {
			dirs = append(dirs, currentPath+"/")
		}
	}

	return dirs
}

// addDirectory 添加目录到待扫描列表
func (c *Controller) addDirectory(dir string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已扫描
	if c.passedUrls[dir] {
		return false
	}

	// 检查是否在排除列表中
	for _, excluded := range c.config.ExcludeSubdirs {
		matched, _ := regexp.MatchString(excluded, dir)
		if matched {
			logger.Debugf("排除子目录: %s", dir)
			return false
		}
	}

	// 检查是否已在待扫描列表中
	for _, existing := range c.directories {
		if existing == dir {
			return false
		}
	}

	// 添加到待扫描列表
	c.directories = append(c.directories, dir)
	logger.Infof("添加递归目录: %s", dir)

	// 添加到字典的extra队列
	c.dictionary.AddExtra(dir)

	return true
}

// setBasePath 设置基础路径
func (c *Controller) setBasePath(url string) {
	// 从URL中提取路径部分
	parts := strings.SplitN(url, "/", 4)
	if len(parts) >= 4 {
		c.basePath = "/" + parts[3]
	} else {
		c.basePath = "/"
	}
}


// handleSignals 处理信号（暂停/恢复）
func (c *Controller) handleSignals(sigChan chan os.Signal) {
	for {
		select {
		case <-sigChan:
			c.handlePause()
		case <-c.ctx.Done():
			return
		}
	}
}

// handlePause 处理暂停
func (c *Controller) handlePause() {
	logger.Warn("检测到中断信号，正在暂停...")

	// 取消上下文以暂停扫描
	c.cancel()

	// 显示交互式菜单
	c.showPauseMenu()
}

// showPauseMenu 显示暂停菜单
func (c *Controller) showPauseMenu() {
	// 第一次尝试读取输入，如果立即失败或为空，说明是非交互环境
	fmt.Println("\n[dirsearch] 暂停菜单")
	fmt.Println("[q]uit    - 退出并保存")
	fmt.Println("[c]ontinue - 继续扫描")
	fmt.Println("[n]ext    - 下一个目标")
	fmt.Print("\n请选择操作: ")

	var choice string
	n, err := fmt.Scanln(&choice)

	// 如果无法读取或读取为空，说明是非交互环境
	if err != nil || n == 0 || strings.TrimSpace(choice) == "" {
		logger.Warn("非交互式环境或无输入，自动退出")
		if c.config.SessionFile != "" {
			logger.Infof("保存会话到: %s", c.config.SessionFile)
			if err := c.SaveSession(c.config.SessionFile); err != nil {
				logger.Errorf("保存会话失败: %v", err)
			}
		}
		logger.Info("退出扫描")
		os.Exit(0)
	}

	// 处理用户输入
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "q", "quit":
		// 退出并保存
		if c.config.SessionFile != "" {
			logger.Infof("保存会话到: %s", c.config.SessionFile)
			if err := c.SaveSession(c.config.SessionFile); err != nil {
				logger.Errorf("保存会话失败: %v", err)
			}
		}
		logger.Info("退出扫描")
		os.Exit(0)

	case "c", "continue":
		// 继续扫描
		logger.Info("继续扫描...")
		return

	case "n", "next":
		// 下一个目标
		logger.Info("跳到下一个目标...")
		return

	default:
		logger.Warn("无效选择，请重试")
		c.showPauseMenuRecursive()
	}
}

// showPauseMenuRecursive 递归版本的暂停菜单（用于重试）
func (c *Controller) showPauseMenuRecursive() {
	for {
		fmt.Println("\n[dirsearch] 暂停菜单")
		fmt.Println("[q]uit    - 退出并保存")
		fmt.Println("[c]ontinue - 继续扫描")
		fmt.Println("[n]ext    - 下一个目标")

		fmt.Print("\n请选择操作: ")
		var choice string
		fmt.Scanln(&choice)

		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "q", "quit":
			if c.config.SessionFile != "" {
				logger.Infof("保存会话到: %s", c.config.SessionFile)
				if err := c.SaveSession(c.config.SessionFile); err != nil {
					logger.Errorf("保存会话失败: %v", err)
				}
			}
			logger.Info("退出扫描")
			os.Exit(0)

		case "c", "continue":
			logger.Info("继续扫描...")
			return

		case "n", "next":
			logger.Info("跳到下一个目标...")
			return

		default:
			logger.Warn("无效选择，请重试")
		}
	}
}

// isTerminal 检查是否在终端中运行
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// checkCrawl 检查是否需要爬取
func (c *Controller) checkCrawl(result *fuzzer.ScanResult) {
	// 只爬取200和301/302响应
	if result.Status != 200 && result.Status != 301 && result.Status != 302 {
		return
	}

	// TODO: 实现完整的爬虫逻辑
	// 需要从response获取HTML内容并解析
	// 暂时记录日志
	logger.Debugf("爬虫检查: %s (状态码: %d)", result.Path, result.Status)

	// 实际爬虫实现需要：
	// 1. 重新请求获取完整响应内容
	// 2. 使用crawler.Crawler解析HTML/robots.txt/纯文本
	// 3. 将发现的路径添加到字典
	//
	// 示例代码：
	// crawler := crawler.NewCrawler()
	// paths := crawler.Crawl(result.URL, c.basePath, content, "html")
	// for _, path := range paths {
	//     c.dictionary.AddExtra(path)
	// }
}
