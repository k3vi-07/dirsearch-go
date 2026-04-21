package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 主配置结构
type Config struct {
	// 基础配置
	URLs          []string `yaml:"urls" mapstructure:"urls"`
	URLList       string   `yaml:"url-list" mapstructure:"url-list"`
	ThreadCount   int      `yaml:"threads" mapstructure:"threads"`
	MaxTime       int      `yaml:"max-time" mapstructure:"max-time"`
	MaxErrors     int      `yaml:"max-errors" mapstructure:"max-errors"`

	// HTTP 配置
	HTTPMethod    string            `yaml:"http-method" mapstructure:"http-method"`
	Headers       map[string]string `yaml:"headers" mapstructure:"headers"`
	HeaderList    string            `yaml:"header-list" mapstructure:"header-list"`
	Data          string            `yaml:"data" mapstructure:"data"`
	DataFile      string            `yaml:"data-file" mapstructure:"data-file"`
	UserAgent     string            `yaml:"user-agent" mapstructure:"user-agent"`
	RandomAgents  bool              `yaml:"random-agents" mapstructure:"random-agents"`
	Cookies       string            `yaml:"cookies" mapstructure:"cookies"`
	Delay         int               `yaml:"delay" mapstructure:"delay"`
	Timeout       int               `yaml:"timeout" mapstructure:"timeout"`

	// 认证配置
	AuthType      string `yaml:"auth-type" mapstructure:"auth-type"`
	Auth          string `yaml:"auth" mapstructure:"auth"`

	// 代理配置
	Proxy         string   `yaml:"proxy" mapstructure:"proxy"`
	ProxyList     string   `yaml:"proxy-list" mapstructure:"proxy-list"`
	ProxyAuth     string   `yaml:"proxy-auth" mapstructure:"proxy-auth"`
	MaxRate       int      `yaml:"max-rate" mapstructure:"max-rate"`
	MaxRetries    int      `yaml:"max-retries" mapstructure:"max-retries"`
	Tor           bool     `yaml:"tor" mapstructure:"tor"`
	NetworkIf     string   `yaml:"network-interface" mapstructure:"network-interface"`

	// 扫描配置
	Extensions        []string `yaml:"extensions" mapstructure:"extensions"`
	ForceExtensions   bool     `yaml:"force-extensions" mapstructure:"force-extensions"`
	ExcludeExtensions []string `yaml:"exclude-extensions" mapstructure:"exclude-extensions"`
	Prefixes          []string `yaml:"prefixes" mapstructure:"prefixes"`
	Suffixes          []string `yaml:"suffixes" mapstructure:"suffixes"`
	Lowercase         bool     `yaml:"lowercase" mapstructure:"lowercase"`
	Uppercase         bool     `yaml:"uppercase" mapstructure:"uppercase"`
	Capitalization    bool     `yaml:"capitalization" mapstructure:"capitalization"`

	// 过滤器配置
	IncludeStatusCodes []string `yaml:"include-status" mapstructure:"include-status"`
	ExcludeStatusCodes []string `yaml:"exclude-status" mapstructure:"exclude-status"`
	ExcludeSizes       []string `yaml:"exclude-sizes" mapstructure:"exclude-sizes"`
	ExcludeTexts       []string `yaml:"exclude-texts" mapstructure:"exclude-texts"`
	ExcludeRegex       string   `yaml:"exclude-regex" mapstructure:"exclude-regex"`
	ExcludeRedirect    string   `yaml:"exclude-redirect" mapstructure:"exclude-redirect"`
	ExcludeResponse    string   `yaml:"exclude-response" mapstructure:"exclude-response"`
	MinResponseSize    int64    `yaml:"min-response-size" mapstructure:"min-response-size"`
	MaxResponseSize    int64    `yaml:"max-response-size" mapstructure:"max-response-size"`
	FilterThreshold    int      `yaml:"filter-threshold" mapstructure:"filter-threshold"`

	// 字典配置
	Wordlists      []string `yaml:"wordlists" mapstructure:"wordlists"`
	StdinWordlist  bool     `yaml:"stdin-wordlist" mapstructure:"stdin-wordlist"`

	// 扫描配置
	Recursive             bool     `yaml:"recursive" mapstructure:"recursive"`
	DeepRecursive         bool     `yaml:"deep-recursive" mapstructure:"deep-recursive"`
	ForceRecursive        bool     `yaml:"force-recursive" mapstructure:"force-recursive"`
	RecursiveDepth        int      `yaml:"recursive-depth" mapstructure:"recursive-depth"`
	RecursionStatusCodes  []string `yaml:"recursion-status-codes" mapstructure:"recursion-status-codes"`
	ExcludeSubdirs        []string `yaml:"exclude-subdirs" mapstructure:"exclude-subdirs"`
	Crawl                 bool     `yaml:"crawl" mapstructure:"crawl"`
	CrawlDepth            int      `yaml:"crawl-depth" mapstructure:"crawl-depth"`
	CrawlMaxPages         int      `yaml:"crawl-max-pages" mapstructure:"crawl-max-pages"`

	// 输出配置
	OutputFormats  []string `yaml:"output-formats" mapstructure:"output-formats"`
	OutputFile     string   `yaml:"output-file" mapstructure:"output-file"`
	LogFile        string   `yaml:"log-file" mapstructure:"log-file"`
	LogLevel       string   `yaml:"log-level" mapstructure:"log-level"`
	Quiet          bool     `yaml:"quiet" mapstructure:"quiet"`

	// 会话配置
	SessionFile    string `yaml:"session" mapstructure:"session"`
	SaveSession   bool   `yaml:"save-session" mapstructure:"save-session"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		// 尝试从默认位置读取配置
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/dirsearch/")
		v.AddConfigPath("$HOME/.dirsearch")
		v.AddConfigPath(".")

		// 静默读取配置文件（可能不存在）
		_ = v.ReadInConfig()
	}

	// 解析配置
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return cfg, nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 基础配置
	v.SetDefault("threads", 30)
	v.SetDefault("max-time", 0)
	v.SetDefault("max-errors", 25)

	// HTTP 配置
	v.SetDefault("http-method", "GET")
	v.SetDefault("headers", map[string]string{})
	v.SetDefault("timeout", 10)
	v.SetDefault("delay", 0)
	v.SetDefault("max-retries", 1)

	// 认证配置
	v.SetDefault("auth-type", "basic")

	// 代理配置
	v.SetDefault("max-rate", 0)

	// 扫描配置
	v.SetDefault("extensions", []string{"php", "html", "js"})
	v.SetDefault("force-extensions", false)

	// 字典配置
	v.SetDefault("wordlists", []string{})

	// 输出配置
	v.SetDefault("output-formats", []string{"plain"})
	v.SetDefault("log-level", "info")
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证线程数
	if cfg.ThreadCount <= 0 {
		return fmt.Errorf("线程数必须大于 0")
	}
	if cfg.ThreadCount > 200 {
		return fmt.Errorf("线程数不能超过 200")
	}

	// 验证超时
	if cfg.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于 0")
	}

	// 验证 HTTP 方法
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true, "PATCH": true,
	}
	if !validMethods[strings.ToUpper(cfg.HTTPMethod)] {
		return fmt.Errorf("无效的 HTTP 方法: %s", cfg.HTTPMethod)
	}
	cfg.HTTPMethod = strings.ToUpper(cfg.HTTPMethod)

	// 验证认证类型
	validAuthTypes := map[string]bool{
		"basic": true, "digest": true, "ntlm": true,
		"bearer": true, "jwt": true,
	}
	if cfg.AuthType != "" && !validAuthTypes[strings.ToLower(cfg.AuthType)] {
		return fmt.Errorf("无效的认证类型: %s", cfg.AuthType)
	}
	if cfg.AuthType != "" {
		cfg.AuthType = strings.ToLower(cfg.AuthType)
	}

	// 验证日志级别
	validLogLevels := map[string]bool{
		"error": true, "warning": true, "info": true, "debug": true,
	}
	if cfg.LogLevel != "" && !validLogLevels[strings.ToLower(cfg.LogLevel)] {
		return fmt.Errorf("无效的日志级别: %s", cfg.LogLevel)
	}
	if cfg.LogLevel != "" {
		cfg.LogLevel = strings.ToLower(cfg.LogLevel)
	}

	// 验证输出格式
	validFormats := map[string]bool{
		"plain": true, "json": true, "csv": true,
		"html": true, "xml": true, "sqlite": true,
	}
	for _, format := range cfg.OutputFormats {
		if !validFormats[strings.ToLower(format)] {
			return fmt.Errorf("无效的输出格式: %s", format)
		}
	}

	// 验证状态码
	if err := validateStatusCodes(cfg.IncludeStatusCodes); err != nil {
		return fmt.Errorf("include-status 错误: %w", err)
	}
	if err := validateStatusCodes(cfg.ExcludeStatusCodes); err != nil {
		return fmt.Errorf("exclude-status 错误: %w", err)
	}

	// 验证 URL 和 URL 列表
	if len(cfg.URLs) == 0 && cfg.URLList == "" {
		return fmt.Errorf("必须指定 URL (-u) 或 URL 列表文件 (-l)")
	}

	// 验证字典
	if len(cfg.Wordlists) == 0 && !cfg.StdinWordlist {
		return fmt.Errorf("必须指定字典文件 (-w)")
	}

	// 验证文件路径
	if cfg.HeaderList != "" {
		if !fileExists(cfg.HeaderList) {
			return fmt.Errorf("请求头文件不存在: %s", cfg.HeaderList)
		}
	}
	if cfg.DataFile != "" {
		if !fileExists(cfg.DataFile) {
			return fmt.Errorf("请求体文件不存在: %s", cfg.DataFile)
		}
	}
	if cfg.ProxyList != "" {
		if !fileExists(cfg.ProxyList) {
			return fmt.Errorf("代理列表文件不存在: %s", cfg.ProxyList)
		}
	}
	if cfg.URLList != "" {
		if !fileExists(cfg.URLList) {
			return fmt.Errorf("URL 列表文件不存在: %s", cfg.URLList)
		}
	}

	return nil
}

// validateStatusCodes 验证状态码格式
func validateStatusCodes(codes []string) error {
	for _, code := range codes {
		// 支持范围格式 (200-299)
		if strings.Contains(code, "-") {
			parts := strings.Split(code, "-")
			if len(parts) != 2 {
				return fmt.Errorf("无效的状态码范围: %s", code)
			}
			// TODO: 验证范围是否有效
		} else {
			// 验证单个状态码
			var sc int
			if _, err := fmt.Sscanf(code, "%d", &sc); err != nil || sc < 100 || sc > 599 {
				return fmt.Errorf("无效的状态码: %s", code)
			}
		}
	}
	return nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetDefaultWordlistPath 获取默认字典路径
func GetDefaultWordlistPath() string {
	home, _ := os.UserHomeDir()
	possiblePaths := []string{
		filepath.Join(home, ".dirsearch", "db", "wordlists", "common.txt"),
		filepath.Join("/usr", "share", "dirb", "wordlists", "common.txt"),
		filepath.Join("/usr", "share", "wordlists", "dirb", "common.txt"),
		"db" + string(filepath.Separator) + "common.txt",
	}

	for _, path := range possiblePaths {
		if fileExists(path) {
			return path
		}
	}

	return ""
}

// ParseStatusCodes 解析状态码字符串为整数列表
func ParseStatusCodes(codes []string) ([]int, error) {
	var result []int

	for _, code := range codes {
		// 支持范围格式 (200-299)
		if strings.Contains(code, "-") {
			parts := strings.Split(code, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("无效的状态码范围: %s", code)
			}

			start, err := parseStatusCode(parts[0])
			if err != nil {
				return nil, err
			}
			end, err := parseStatusCode(parts[1])
			if err != nil {
				return nil, err
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			sc, err := parseStatusCode(code)
			if err != nil {
				return nil, err
			}
			result = append(result, sc)
		}
	}

	return result, nil
}

// parseStatusCode 解析单个状态码
func parseStatusCode(code string) (int, error) {
	var sc int
	if _, err := fmt.Sscanf(code, "%d", &sc); err != nil {
		return 0, fmt.Errorf("无效的状态码: %s", code)
	}
	if sc < 100 || sc > 599 {
		return 0, fmt.Errorf("状态码超出范围: %s", code)
	}
	return sc, nil
}

// GetTimeoutDuration 获取超时时间
func (c *Config) GetTimeoutDuration() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetDelayDuration 获取延迟时间
func (c *Config) GetDelayDuration() time.Duration {
	return time.Duration(c.Delay) * time.Millisecond
}

// GetMaxTimeDuration 获取最大扫描时间
func (c *Config) GetMaxTimeDuration() time.Duration {
	if c.MaxTime <= 0 {
		return 0 // 无限制
	}
	return time.Duration(c.MaxTime) * time.Second
}

// HasExtension 检查是否有扩展名
func (c *Config) HasExtension() bool {
	return len(c.Extensions) > 0
}

// IsForceExtension 是否强制扩展名
func (c *Config) IsForceExtension() bool {
	return c.ForceExtensions
}

// GetOutputFile 获取输出文件路径
func (c *Config) GetOutputFile() string {
	if c.OutputFile != "" {
		return c.OutputFile
	}
	return ""
}

// NeedsAuth 是否需要认证
func (c *Config) NeedsAuth() bool {
	return c.Auth != "" && c.AuthType != ""
}
