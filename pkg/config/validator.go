package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

// Validator 配置验证器
type Validator struct {
	errors []string
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// Validate 验证配置
func (v *Validator) Validate(cfg *Config) error {
	v.errors = make([]string, 0)

	// 验证 URL
	if err := v.validateURLs(cfg); err != nil {
		return err
	}

	// 验证线程数
	v.validateThreads(cfg.ThreadCount)

	// 验证超时
	v.validateTimeout(cfg.Timeout)

	// 验证 HTTP 方法
	v.validateHTTPMethod(cfg.HTTPMethod)

	// 验证认证
	v.validateAuth(cfg.AuthType, cfg.Auth)

	// 验证代理
	if err := v.validateProxy(cfg.Proxy); err != nil {
		return err
	}

	// 验证状态码
	if err := v.validateStatusCodeList(cfg.IncludeStatusCodes); err != nil {
		return err
	}
	if err := v.validateStatusCodeList(cfg.ExcludeStatusCodes); err != nil {
		return err
	}

	// 验证响应大小
	if err := v.validateResponseSizes(cfg.ExcludeSizes); err != nil {
		return err
	}

	// 验证字典
	if err := v.validateWordlists(cfg.Wordlists, cfg.StdinWordlist); err != nil {
		return err
	}

	// 验证输出格式
	if err := v.validateOutputFormats(cfg.OutputFormats); err != nil {
		return err
	}

	// 验证文件路径
	v.validateFilePaths(cfg)

	// 验证日志级别
	v.validateLogLevel(cfg.LogLevel)

	if len(v.errors) > 0 {
		return fmt.Errorf("配置验证失败:\n- %s", strings.Join(v.errors, "\n- "))
	}

	return nil
}

// validateURLs 验证 URL
func (v *Validator) validateURLs(cfg *Config) error {
	// 检查是否至少有一个 URL 源
	if len(cfg.URLs) == 0 && cfg.URLList == "" {
		v.errors = append(v.errors, "必须指定 URL (-u) 或 URL 列表文件 (-l)")
		return fmt.Errorf("缺少 URL")
	}

	// 验证每个 URL
	for _, u := range cfg.URLs {
		if _, err := url.Parse(u); err != nil {
			v.errors = append(v.errors, fmt.Sprintf("无效的 URL: %s", u))
		}
	}

	// 验证 URL 列表文件
	if cfg.URLList != "" {
		if !fileExists(cfg.URLList) {
			v.errors = append(v.errors, fmt.Sprintf("URL 列表文件不存在: %s", cfg.URLList))
		}
	}

	return nil
}

// validateThreads 验证线程数
func (v *Validator) validateThreads(threads int) {
	if threads <= 0 {
		v.errors = append(v.errors, "线程数必须大于 0")
	}
	if threads > 200 {
		v.errors = append(v.errors, "线程数不能超过 200")
	}
}

// validateTimeout 验证超时时间
func (v *Validator) validateTimeout(timeout int) {
	if timeout <= 0 {
		v.errors = append(v.errors, "超时时间必须大于 0")
	}
	if timeout > 300 {
		v.errors = append(v.errors, "超时时间不能超过 300 秒")
	}
}

// validateHTTPMethod 验证 HTTP 方法
func (v *Validator) validateHTTPMethod(method string) {
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true, "PATCH": true, "TRACE": true,
	}

	if !validMethods[strings.ToUpper(method)] {
		v.errors = append(v.errors, fmt.Sprintf("无效的 HTTP 方法: %s", method))
	}
}

// validateAuth 验证认证配置
func (v *Validator) validateAuth(authType, auth string) {
	if auth == "" {
		return
	}

	validTypes := map[string]bool{
		"basic": true, "digest": true, "ntlm": true,
		"bearer": true, "jwt": true,
	}

	authTypeLower := strings.ToLower(authType)
	if !validTypes[authTypeLower] {
		v.errors = append(v.errors, fmt.Sprintf("无效的认证类型: %s", authType))
		return
	}

	// 验证认证凭据格式
	switch authTypeLower {
	case "basic", "digest", "ntlm":
		if !strings.Contains(auth, ":") {
			v.errors = append(v.errors, fmt.Sprintf("%s 认证需要 username:password 格式", authType))
		}
	case "bearer", "jwt":
		if len(auth) < 10 {
			v.errors = append(v.errors, fmt.Sprintf("%s token 太短", authType))
		}
	}
}

// validateProxy 验证代理配置
func (v *Validator) validateProxy(proxy string) error {
	if proxy == "" {
		return nil
	}

	// 验证代理 URL 格式
	u, err := url.Parse(proxy)
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("无效的代理 URL: %s", proxy))
		return err
	}

	// 验证代理协议
	validSchemes := map[string]bool{
		"http": true, "https": true, "socks5": true,
	}
	if !validSchemes[u.Scheme] {
		v.errors = append(v.errors, fmt.Sprintf("不支持的代理协议: %s", u.Scheme))
	}

	return nil
}

// validateStatusCodeList 验证状态码列表
func (v *Validator) validateStatusCodeList(codes []string) error {
	for _, code := range codes {
		if strings.Contains(code, "-") {
			// 范围格式 (200-299)
			parts := strings.Split(code, "-")
			if len(parts) != 2 {
				v.errors = append(v.errors, fmt.Sprintf("无效的状态码范围: %s", code))
				continue
			}

			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])

			if err1 != nil || err2 != nil {
				v.errors = append(v.errors, fmt.Sprintf("无效的状态码范围: %s", code))
				continue
			}

			if start < 100 || start > 599 || end < 100 || end > 599 || start > end {
				v.errors = append(v.errors, fmt.Sprintf("无效的状态码范围: %s", code))
			}
		} else {
			// 单个状态码
			sc, err := strconv.Atoi(code)
			if err != nil {
				v.errors = append(v.errors, fmt.Sprintf("无效的状态码: %s", code))
				continue
			}

			if sc < 100 || sc > 599 {
				v.errors = append(v.errors, fmt.Sprintf("状态码超出范围: %s", code))
			}
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("状态码验证失败")
	}

	return nil
}

// validateResponseSizes 验证响应大小
func (v *Validator) validateResponseSizes(sizes []string) error {
	for _, size := range sizes {
		// 支持格式: 100, 100B, 100KB, 100MB, 100GB
		size = strings.ToUpper(strings.TrimSpace(size))

		var numStr, unit string
		if idx := strings.IndexAny(size, "KMGT"); idx != -1 {
			numStr = size[:idx]
			unit = size[idx:]
		} else {
			numStr = size
		}

		num, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			v.errors = append(v.errors, fmt.Sprintf("无效的响应大小: %s", size))
			continue
		}

		if num <= 0 {
			v.errors = append(v.errors, fmt.Sprintf("响应大小必须大于 0: %s", size))
			continue
		}

		// 验证单位
		if unit != "" && unit != "B" && unit != "KB" && unit != "MB" && unit != "GB" {
			v.errors = append(v.errors, fmt.Sprintf("无效的响应大小单位: %s", unit))
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("响应大小验证失败")
	}

	return nil
}

// validateWordlists 验证字典文件
func (v *Validator) validateWordlists(wordlists []string, useStdin bool) error {
	if useStdin {
		return nil
	}

	if len(wordlists) == 0 {
		v.errors = append(v.errors, "必须指定字典文件 (-w)")
		return fmt.Errorf("缺少字典文件")
	}

	for _, wordlist := range wordlists {
		// 跳过内置字典标记
		if wordlist == "__builtin__" {
			continue
		}
		if !fileExists(wordlist) {
			v.errors = append(v.errors, fmt.Sprintf("字典文件不存在: %s", wordlist))
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("字典文件验证失败")
	}

	return nil
}

// validateOutputFormats 验证输出格式
func (v *Validator) validateOutputFormats(formats []string) error {
	if len(formats) == 0 {
		v.errors = append(v.errors, "必须指定至少一种输出格式")
		return fmt.Errorf("缺少输出格式")
	}

	validFormats := map[string]bool{
		"plain": true, "simple": true, "json": true,
		"csv": true, "xml": true, "html": true,
		"markdown": true, "md": true, "sqlite": true,
	}

	for _, format := range formats {
		formatLower := strings.ToLower(format)
		if !validFormats[formatLower] {
			v.errors = append(v.errors, fmt.Sprintf("无效的输出格式: %s", format))
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("输出格式验证失败")
	}

	return nil
}

// validateFilePaths 验证文件路径
func (v *Validator) validateFilePaths(cfg *Config) {
	// 验证请求头文件
	if cfg.HeaderList != "" && !fileExists(cfg.HeaderList) {
		v.errors = append(v.errors, fmt.Sprintf("请求头文件不存在: %s", cfg.HeaderList))
	}

	// 验证请求体文件
	if cfg.DataFile != "" && !fileExists(cfg.DataFile) {
		v.errors = append(v.errors, fmt.Sprintf("请求体文件不存在: %s", cfg.DataFile))
	}

	// 验证代理列表文件
	if cfg.ProxyList != "" && !fileExists(cfg.ProxyList) {
		v.errors = append(v.errors, fmt.Sprintf("代理列表文件不存在: %s", cfg.ProxyList))
	}

	// 验证输出目录是否存在（如果指定）
	if cfg.OutputFile != "" {
		dir := filepath.Dir(cfg.OutputFile)
		if dir != "." && !fileExists(dir) {
			v.errors = append(v.errors, fmt.Sprintf("输出目录不存在: %s", dir))
		}
	}

	// 验证日志目录是否存在（如果指定）
	if cfg.LogFile != "" {
		dir := filepath.Dir(cfg.LogFile)
		if dir != "." && !fileExists(dir) {
			v.errors = append(v.errors, fmt.Sprintf("日志目录不存在: %s", dir))
		}
	}
}

// validateLogLevel 验证日志级别
func (v *Validator) validateLogLevel(logLevel string) {
	if logLevel == "" {
		return
	}

	validLevels := map[string]bool{
		"error": true, "warning": true, "warn": true,
		"info": true, "debug": true, "trace": true,
	}

	if !validLevels[strings.ToLower(logLevel)] {
		v.errors = append(v.errors, fmt.Sprintf("无效的日志级别: %s", logLevel))
	}
}

// GetErrors 获取所有错误
func (v *Validator) GetErrors() []string {
	return v.errors
}

// HasErrors 是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}
