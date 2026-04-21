package main

import (
	"fmt"
	"os"

	"github.com/youruser/dirsearch-go/internal/logger"
	"github.com/youruser/dirsearch-go/pkg/config"
	"github.com/youruser/dirsearch-go/pkg/controller"
	"github.com/spf13/cobra"
)

var (
	// Version 版本号
	Version = "dev"
	// BuildTime 编译时间
	BuildTime = "unknown"
)

// 配置变量
var cfg config.Config

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dirsearch",
		Short: "Web路径扫描工具 - 轻量级目录爆破工具",
		Long: `dirsearch-go - 高性能Web路径扫描工具

使用Go重写的dirsearch，性能更优，单文件部署
支持多线程、并发扫描、智能过滤等功能`,
		Example: `  # 扫描单个目标（使用内置字典）
  dirsearch -u http://example.com

  # 使用自定义字典
  dirsearch -u http://example.com -w db/common.txt

  # 批量扫描
  dirsearch -l urls.txt -t 50 -f json,html

  # 指定扩展名和代理
  dirsearch -u http://example.com -x php,html,js --proxy http://127.0.0.1:8080

  # 递归扫描
  dirsearch -u http://example.com -r --depth 3`,
		Run: run,
	}

	// 基础选项
	rootCmd.Flags().StringSliceVarP(&cfg.URLs, "url", "u", []string{}, "目标URL")
	rootCmd.Flags().StringVarP(&cfg.URLList, "url-list", "l", "", "URL列表文件")
	rootCmd.Flags().IntVarP(&cfg.ThreadCount, "threads", "t", 30, "线程数")
	rootCmd.Flags().IntVar(&cfg.MaxTime, "max-time", 0, "最大扫描时间（秒）")
	rootCmd.Flags().IntVar(&cfg.MaxErrors, "max-errors", 25, "最大错误数")

	// HTTP 选项
	rootCmd.Flags().StringVarP(&cfg.HTTPMethod, "method", "m", "GET", "HTTP方法")
	rootCmd.Flags().StringToStringVarP(&cfg.Headers, "header", "H", nil, "HTTP请求头")
	rootCmd.Flags().StringVar(&cfg.HeaderList, "header-list", "", "请求头列表文件")
	rootCmd.Flags().StringVarP(&cfg.Data, "data", "d", "", "HTTP请求数据")
	rootCmd.Flags().StringVar(&cfg.DataFile, "data-file", "", "数据文件")
	rootCmd.Flags().StringVarP(&cfg.UserAgent, "user-agent", "a", "", "自定义User-Agent")
	rootCmd.Flags().BoolVar(&cfg.RandomAgents, "random-agents", false, "使用随机User-Agent")
	rootCmd.Flags().StringVar(&cfg.Cookies, "cookies", "", "Cookie")
	rootCmd.Flags().IntVar(&cfg.Delay, "delay", 0, "请求间延迟（毫秒）")
	rootCmd.Flags().IntVar(&cfg.Timeout, "timeout", 10, "请求超时（秒）")

	// 认证选项
	rootCmd.Flags().StringVar(&cfg.AuthType, "auth-type", "basic", "认证类型 (basic/digest/ntlm/bearer/jwt)")
	rootCmd.Flags().StringVar(&cfg.Auth, "auth", "", "认证凭据")

	// 代理选项
	rootCmd.Flags().StringVar(&cfg.Proxy, "proxy", "", "代理地址")
	rootCmd.Flags().StringVar(&cfg.ProxyList, "proxy-list", "", "代理列表文件")
	rootCmd.Flags().StringVar(&cfg.ProxyAuth, "proxy-auth", "", "代理认证")
	rootCmd.Flags().IntVar(&cfg.MaxRate, "max-rate", 0, "最大请求速率（每秒请求数）")
	rootCmd.Flags().IntVar(&cfg.MaxRetries, "max-retries", 1, "最大重试次数")
	rootCmd.Flags().BoolVar(&cfg.Tor, "tor", false, "使用Tor网络")
	rootCmd.Flags().StringVar(&cfg.NetworkIf, "network-interface", "", "网络接口")

	// 字典选项
	rootCmd.Flags().StringSliceVarP(&cfg.Wordlists, "wordlists", "w", []string{}, "字典文件（不指定则使用内置9680条字典）")
	rootCmd.Flags().BoolVar(&cfg.StdinWordlist, "stdin-wordlist", false, "从标准输入读取字典")
	rootCmd.Flags().StringSliceVarP(&cfg.Extensions, "extensions", "x", []string{"php", "html", "js"}, "扩展名")
	rootCmd.Flags().BoolVar(&cfg.ForceExtensions, "force-extensions", false, "强制扩展名")
	rootCmd.Flags().StringSliceVar(&cfg.ExcludeExtensions, "exclude-extensions", []string{}, "排除扩展名")
	rootCmd.Flags().StringSliceVar(&cfg.Prefixes, "prefixes", []string{}, "自定义前缀")
	rootCmd.Flags().StringSliceVar(&cfg.Suffixes, "suffixes", []string{}, "自定义后缀")
	rootCmd.Flags().BoolVar(&cfg.Lowercase, "lowercase", false, "转小写")
	rootCmd.Flags().BoolVar(&cfg.Uppercase, "uppercase", false, "转大写")
	rootCmd.Flags().BoolVar(&cfg.Capitalization, "capitalization", false, "首字母大写")

	// 过滤器选项
	rootCmd.Flags().StringSliceVar(&cfg.IncludeStatusCodes, "include-status", []string{}, "包含状态码")
	rootCmd.Flags().StringSliceVar(&cfg.ExcludeStatusCodes, "exclude-status", []string{"400", "403", "404"}, "排除状态码")
	rootCmd.Flags().StringSliceVar(&cfg.ExcludeSizes, "exclude-size", []string{}, "排除响应大小")
	rootCmd.Flags().StringSliceVar(&cfg.ExcludeTexts, "exclude-texts", []string{}, "排除响应文本")
	rootCmd.Flags().StringVar(&cfg.ExcludeRegex, "exclude-regex", "", "排除正则表达式")
	rootCmd.Flags().StringVar(&cfg.ExcludeRedirect, "exclude-redirect", "", "排除重定向模式")
	rootCmd.Flags().StringVar(&cfg.ExcludeResponse, "exclude-response", "", "自定义排除响应")
	rootCmd.Flags().Int64Var(&cfg.MinResponseSize, "min-response-size", 0, "最小响应大小")
	rootCmd.Flags().Int64Var(&cfg.MaxResponseSize, "max-response-size", 0, "最大响应大小")
	rootCmd.Flags().IntVar(&cfg.FilterThreshold, "filter-threshold", 0, "过滤阈值")

	// 扫描选项
	rootCmd.Flags().BoolVar(&cfg.Recursive, "recursive", false, "递归扫描")
	rootCmd.Flags().BoolVar(&cfg.DeepRecursive, "deep-recursive", false, "深度递归扫描")
	rootCmd.Flags().BoolVar(&cfg.ForceRecursive, "force-recursive", false, "强制递归扫描")
	rootCmd.Flags().IntVar(&cfg.RecursiveDepth, "recursive-depth", 3, "递归深度")
	rootCmd.Flags().StringSliceVar(&cfg.RecursionStatusCodes, "recursion-status-codes", []string{"200", "204", "301", "302", "307", "401", "403"}, "触发递归的状态码")
	rootCmd.Flags().StringSliceVar(&cfg.ExcludeSubdirs, "exclude-subdirs", []string{}, "排除的子目录")
	rootCmd.Flags().BoolVar(&cfg.Crawl, "crawl", false, "启用爬虫")
	rootCmd.Flags().IntVar(&cfg.CrawlDepth, "crawl-depth", 2, "爬虫深度")
	rootCmd.Flags().IntVar(&cfg.CrawlMaxPages, "crawl-max-pages", 50, "爬虫最大页面数")

	// 输出选项
	rootCmd.Flags().StringSliceVarP(&cfg.OutputFormats, "format", "f", []string{"plain"}, "输出格式 (plain/json/csv/xml/html/markdown)")
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "输出文件")
	rootCmd.Flags().StringVar(&cfg.LogFile, "log-file", "", "日志文件")
	rootCmd.Flags().StringVar(&cfg.LogLevel, "log-level", "info", "日志级别 (debug/info/warn/error)")
	rootCmd.Flags().BoolVarP(&cfg.Quiet, "quiet", "q", false, "安静模式（仅显示结果）")

	// 会话选项
	rootCmd.Flags().StringVar(&cfg.SessionFile, "session", "", "会话文件")
	rootCmd.Flags().BoolVar(&cfg.SaveSession, "save-session", false, "保存会话")

	// 版本
	rootCmd.Flags().Bool("version", false, "显示版本信息")

	cobra.CheckErr(rootCmd.Execute())
}

func run(cmd *cobra.Command, args []string) {
	// 检查是否请求版本
	if version, _ := cmd.Flags().GetBool("version"); version {
		fmt.Printf("dirsearch-go v%s (编译时间: %s)\n", Version, BuildTime)
		return
	}

	// 检查是否没有URL参数，显示帮助
	urls, _ := cmd.Flags().GetStringSlice("url")
	urlList, _ := cmd.Flags().GetString("url-list")
	if len(urls) == 0 && urlList == "" {
		cmd.Help()
		return
	}

	// 打印启动信息
	printBanner()

	// 初始化日志系统
	if err := logger.Init(cfg.LogLevel, cfg.LogFile); err != nil {
		fmt.Fprintf(os.Stderr, "日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	logger.Info("dirsearch-go 启动中...")

	// 加载配置
	loadedCfg, err := config.LoadConfig("")
	if err != nil {
		logger.Errorf("配置加载失败: %v", err)
		os.Exit(1)
	}

	// 合并命令行参数
	mergeConfig(loadedCfg, &cfg)

	// 字典处理：如果未指定，使用内置默认字典（在验证之前）
	if len(loadedCfg.Wordlists) == 0 {
		logger.Info("未指定字典，使用内置默认字典 (9680条路径)")
		loadedCfg.Wordlists = []string{"__builtin__"} // 标记使用内置字典
	}

	// 验证配置
	validator := config.NewValidator()
	if err := validator.Validate(loadedCfg); err != nil {
		logger.Errorf("配置验证失败: %v", err)
		os.Exit(1)
	}

	logger.Info("配置加载成功")
	logger.Infof("URLs: %v", loadedCfg.URLs)
	logger.Infof("Threads: %d", loadedCfg.ThreadCount)
	logger.Infof("Extensions: %v", loadedCfg.Extensions)
	if len(loadedCfg.Wordlists) > 0 && loadedCfg.Wordlists[0] != "__builtin__" {
		logger.Infof("Wordlists: %v", loadedCfg.Wordlists)
	}

	// 创建控制器
	ctrl, err := controller.NewController(loadedCfg)
	if err != nil {
		logger.Errorf("创建控制器失败: %v", err)
		os.Exit(1)
	}

	// 运行扫描
	if err := ctrl.Run(); err != nil {
		logger.Errorf("扫描失败: %v", err)
		os.Exit(1)
	}

	logger.Info("扫描完成!")
}

// printBanner 打印启动横幅
func printBanner() {
	banner := `
  _                          _  __
 | |   _ __   ___  __   __ (_) / _  ___
 | |  | '_ \ / _ \ \ \ / / | | || |/ _ \
 | |__| | | |  __/  \ V /  | | || | (_) |
 |____|_| |_|\___|   \_/   |_|_||_|\___/
 by dirsearch-go v%s (built: %s)
`
	fmt.Printf(banner, Version, BuildTime)
	fmt.Println()
}

// mergeConfig 合并命令行配置到加载的配置
func mergeConfig(loaded *config.Config, cmdCfg *config.Config) {
	// URL
	if len(cmdCfg.URLs) > 0 {
		loaded.URLs = cmdCfg.URLs
	}
	if cmdCfg.URLList != "" {
		loaded.URLList = cmdCfg.URLList
	}

	// 线程
	if cmdCfg.ThreadCount != 30 { // 非默认值
		loaded.ThreadCount = cmdCfg.ThreadCount
	}

	// HTTP 方法
	if cmdCfg.HTTPMethod != "GET" { // 非默认值
		loaded.HTTPMethod = cmdCfg.HTTPMethod
	}

	// Headers
	if len(cmdCfg.Headers) > 0 {
		if loaded.Headers == nil {
			loaded.Headers = make(map[string]string)
		}
		for k, v := range cmdCfg.Headers {
			loaded.Headers[k] = v
		}
	}

	// User-Agent
	if cmdCfg.UserAgent != "" {
		loaded.UserAgent = cmdCfg.UserAgent
	}

	// 认证
	if cmdCfg.Auth != "" {
		loaded.Auth = cmdCfg.Auth
		loaded.AuthType = cmdCfg.AuthType
	}

	// 代理
	if cmdCfg.Proxy != "" {
		loaded.Proxy = cmdCfg.Proxy
	}

	// 扩展名
	if !equalStringSlice(cmdCfg.Extensions, []string{"php", "html", "js"}) { // 非默认值
		loaded.Extensions = cmdCfg.Extensions
	}

	// 字典
	if len(cmdCfg.Wordlists) > 0 {
		loaded.Wordlists = cmdCfg.Wordlists
	}

	// 输出
	if !equalStringSlice(cmdCfg.OutputFormats, []string{"plain"}) { // 非默认值
		loaded.OutputFormats = cmdCfg.OutputFormats
	}
	if cmdCfg.OutputFile != "" {
		loaded.OutputFile = cmdCfg.OutputFile
	}
	if cmdCfg.Quiet { // 安静模式
		loaded.Quiet = cmdCfg.Quiet
	}

	// 过滤器
	if len(cmdCfg.ExcludeStatusCodes) > 0 {
		loaded.ExcludeStatusCodes = cmdCfg.ExcludeStatusCodes
	}
	if len(cmdCfg.IncludeStatusCodes) > 0 {
		loaded.IncludeStatusCodes = cmdCfg.IncludeStatusCodes
	}
	if len(cmdCfg.ExcludeSizes) > 0 {
		loaded.ExcludeSizes = cmdCfg.ExcludeSizes
	}
	if len(cmdCfg.ExcludeTexts) > 0 {
		loaded.ExcludeTexts = cmdCfg.ExcludeTexts
	}
	if cmdCfg.ExcludeRegex != "" {
		loaded.ExcludeRegex = cmdCfg.ExcludeRegex
	}
	if cmdCfg.ExcludeRedirect != "" {
		loaded.ExcludeRedirect = cmdCfg.ExcludeRedirect
	}
	if cmdCfg.ExcludeResponse != "" {
		loaded.ExcludeResponse = cmdCfg.ExcludeResponse
	}
	if cmdCfg.MinResponseSize > 0 {
		loaded.MinResponseSize = cmdCfg.MinResponseSize
	}
	if cmdCfg.MaxResponseSize > 0 {
		loaded.MaxResponseSize = cmdCfg.MaxResponseSize
	}

	// 扫描选项
	if cmdCfg.Recursive {
		loaded.Recursive = true
	}
	if cmdCfg.RecursiveDepth != 3 {
		loaded.RecursiveDepth = cmdCfg.RecursiveDepth
	}
	if cmdCfg.Crawl {
		loaded.Crawl = true
	}
	if cmdCfg.CrawlDepth != 2 {
		loaded.CrawlDepth = cmdCfg.CrawlDepth
	}
	if cmdCfg.CrawlMaxPages != 50 {
		loaded.CrawlMaxPages = cmdCfg.CrawlMaxPages
	}

	// 字典选项
	if len(cmdCfg.Prefixes) > 0 {
		loaded.Prefixes = cmdCfg.Prefixes
	}
	if len(cmdCfg.Suffixes) > 0 {
		loaded.Suffixes = cmdCfg.Suffixes
	}
	if cmdCfg.ForceExtensions {
		loaded.ForceExtensions = true
	}
	if len(cmdCfg.ExcludeExtensions) > 0 {
		loaded.ExcludeExtensions = cmdCfg.ExcludeExtensions
	}
	if cmdCfg.Lowercase {
		loaded.Lowercase = true
	}
	if cmdCfg.Uppercase {
		loaded.Uppercase = true
	}
	if cmdCfg.Capitalization {
		loaded.Capitalization = true
	}

	// 日志
	if cmdCfg.LogLevel != "info" {
		loaded.LogLevel = cmdCfg.LogLevel
	}
	if cmdCfg.LogFile != "" {
		loaded.LogFile = cmdCfg.LogFile
	}

	// 会话
	if cmdCfg.SessionFile != "" {
		loaded.SessionFile = cmdCfg.SessionFile
	}
	if cmdCfg.SaveSession {
		loaded.SaveSession = true
	}
}

// equalStringSlice 比较字符串切片是否相等
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
