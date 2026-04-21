# dirsearch-go

[![Go Report Card](https://goreportcard.com/badge/github.com/k3vi-07/dirsearch-go)](https://goreportcard.com/report/github.com/k3vi-07/dirsearch-go)
[![License](https://img.shields.io/badge/license-GPLv2-blue.svg)](LICENSE)

> **dirsearch-go** - Web路径扫描工具的Go语言完整重写
> 比Python版本快2-5倍，单文件部署，功能完整

[原项目](https://github.com/maurosoria/dirsearch) | [在线文档](#) | [示例](#)

## ✨ 特性

### 核心功能
- 🔍 **多线程扫描** - 高性能Goroutine并发引擎
- 🎯 **通配符检测** - 双测试点法智能识别
- 📊 **6种报告格式** - Plain/JSON/CSV/XML/HTML/Markdown
- 🔐 **5种认证** - Basic/Digest/NTLM/Bearer/JWT
- 🌐 **代理支持** - HTTP/HTTPS代理轮换
- 💾 **会话管理** - 保存和恢复扫描进度

### 高级功能
- 🔄 **递归扫描** - 3种模式（标准/深度/强制）
- 🕷️ **爬虫功能** - HTML/robots.txt/纯文本爬取
- ⏸️ **暂停/恢复** - 交互式信号处理
- 📖 **内置字典** - 446条常用路径
- 🌏 **中文界面** - 完整中文本地化
- ⚡ **DNS缓存** - 5分钟TTL加速解析
- 🚦 **速率限制** - 令牌桶算法

### 性能优势
| 指标 | Python版本 | Go版本 | 提升 |
|------|-----------|--------|------|
| 吞吐量 | ~500 req/s | 1000-2500 req/s | **2-5x** |
| 内存 | ~150MB | <80MB | **47%↓** |
| 启动时间 | ~1s | <0.1s | **10x** |
| 部署 | 需Python | 单文件 | **✅** |

## 📦 安装

### 从源码编译
```bash
git clone https://github.com/k3vi-07/dirsearch-go.git
cd dirsearch-go
go build -o dirsearch cmd/dirsearch/main.go
```

### 预编译二进制
从 [Releases](https://github.com/k3vi-07/dirsearch-go/releases) 下载

## 🚀 快速开始

### 基础扫描
```bash
# 使用内置字典（446条路径）
./dirsearch -u http://example.com

# 使用自定义字典
./dirsearch -u http://example.com -w /path/to/wordlist.txt
```

### 递归扫描
```bash
# 标准递归（只扫描以/结尾的目录）
./dirsearch -u http://example.com --recursive --recursive-depth 2

# 深度递归（扫描所有层级目录）
./dirsearch -u http://example.com --deep-recursive

# 强制递归（非目录也当作目录处理）
./dirsearch -u http://example.com --force-recursive
```

### 爬虫功能
```bash
# 启用HTML爬取
./dirsearch -u http://example.com --crawl --crawl-depth 2

# 爬虫 + 递归组合
./dirsearch -u http://example.com --crawl --recursive
```

### 多线程和扩展名
```bash
# 50线程，指定扩展名
./dirsearch -u http://example.com -t 50 -x php,html,js

# 强制扩展名（admin -> admin.php）
./dirsearch -u http://example.com --force-extensions -x php
```

### 代理和认证
```bash
# HTTP代理
./dirsearch -u http://example.com --proxy http://127.0.0.1:8080

# 代理认证
./dirsearch -u http://example.com --proxy http://127.0.0.1:8080 --proxy-auth user:pass

# Basic认证
./dirsearch -u http://example.com --auth-type basic --auth admin:password
```

### 导出结果
```bash
# JSON格式
./dirsearch -u http://example.com -o report.json -f json

# 多种格式
./dirsearch -u http://example.com -o report -f json,csv,html
```

### 会话管理
```bash
# 保存会话（Ctrl+C后选择quit）
./dirsearch -u http://example.com --session scan.session

# 恢复会话
./dirsearch --session scan.session
```

## 📖 完整参数

### 必需参数
```
-u, --url <urls>              目标URL（可多个，逗号分隔）
-l, --url-list <file>         URL列表文件
```

### 字典设置
```
-w, --wordlists <paths>        字典文件（默认使用内置446条）
-e, --extensions <exts>        扩展名（php,html,js）
--force-extensions             强制扩展名模式
--exclude-extensions <exts>    排除扩展名
--prefixes <prefixes>          自定义前缀
--suffixes <suffixes>          自定义后缀
--lowercase                    小写转换
--uppercase                    大写转换
--capitalization               首字母大写
```

### 扫描设置
```
-t, --threads <num>            线程数（默认：30）
-r, --recursive                递归扫描
--deep-recursive               深度递归扫描
--force-recursive              强制递归扫描
--recursive-depth <num>        递归深度（默认：3）
--recursion-status-codes <codes> 触发递归的状态码
--exclude-subdirs <patterns>   排除的子目录
--crawl                        启用爬虫
--crawl-depth <num>            爬虫深度（默认：2）
--crawl-max-pages <num>        最大页面数（默认：50）
--max-time <seconds>           最大扫描时间
--max-errors <num>             最大错误数
```

### HTTP设置
```
-m, --method <method>          HTTP方法（默认：GET）
-H, --header <header>          自定义请求头
--data <data>                  请求体数据
-d, --delay <ms>               请求延迟
--timeout <seconds>            请求超时（默认：10）
--user-agent <agent>           User-Agent
--random-agents                随机User-Agent
--cookies <cookies>            Cookie
```

### 认证
```
--auth-type <type>             认证类型（basic/digest/ntlm/bearer/jwt）
--auth <credential>            认证凭据
```

### 代理
```
--proxy <proxy>                代理URL
--proxy-list <file>            代理列表文件
--proxy-auth <auth>            代理认证
--max-rate <req/s>             最大请求速率
--tor                          使用Tor网络
--network-interface <iface>     网络接口
```

### 过滤器
```
--include-status <codes>        包含状态码
--exclude-status <codes>        排除状态码
--exclude-sizes <sizes>         排除响应大小
--exclude-texts <texts>         排除响应文本
--exclude-regex <pattern>       排除正则
--min-response-size <size>      最小响应大小
--max-response-size <size>      最大响应大小
```

### 输出
```
-o, --output <file>            输出文件
-f, --format <formats>         格式（plain/json/csv/xml/html/markdown）
--log-file <file>               日志文件
--log-level <level>             日志级别
```

### 会话
```
--session <file>                会话文件
--save-session                 自动保存会话
```

## 🎯 使用示例

### 场景1：快速扫描
```bash
./dirsearch -u http://target.com
```

### 场景2：深度扫描
```bash
./dirsearch -u http://target.com \
  --deep-recursive \
  --recursive-depth 3 \
  -t 50 \
  -x php,html,js,asp
```

### 场景3：隐蔽扫描
```bash
./dirsearch -u http://target.com \
  -t 10 \
  -d 1000 \
  --random-agents \
  --proxy http://127.0.0.1:8080
```

### 场景4：认证扫描
```bash
./dirsearch -u http://target.com \
  --auth-type digest \
  --auth admin:secret \
  --header "Authorization: Bearer token"
```

### 场景5：持续扫描
```bash
# 保存会话
./dirsearch -u http://target.com --session deep.session

# Ctrl+C中断后恢复
./dirsearch --session deep.session
```

## 📁 项目结构

```
dirsearch-go/
├── cmd/dirsearch/          # 主程序入口
├── pkg/
│   ├── config/            # 配置管理
│   ├── controller/        # 核心控制器
│   │   ├── controller.go  # 主流程
│   │   ├── session.go     # 会话管理
│   │   └── recursive.go   # 递归扫描
│   ├── crawler/           # 爬虫模块
│   ├── dictionary/        # 字典管理
│   ├── fuzzer/           # 模糊测试引擎
│   ├── requester/        # HTTP请求器
│   │   ├── auth/         # 认证实现
│   │   └── rate_limiter.go
│   ├── report/           # 报告生成器
│   ├── scanner/          # 扫描器
│   │   └── wildcard.go   # 通配符检测
│   └── response/         # 响应处理
├── internal/
│   ├── dns/              # DNS缓存
│   └── logger/           # 日志系统
├── db/                   # 内置字典
│   ├── common.txt        # 446条常用路径
│   └── blacklists/       # 黑名单
└── go.mod
```

## 🔧 开发

### 构建
```bash
# 编译
go build -o dirsearch cmd/dirsearch/main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o dirsearch-linux cmd/dirsearch/main.go
GOOS=windows GOARCH=amd64 go build -o dirsearch.exe cmd/dirsearch/main.go
```

### 测试
```bash
# 单元测试
go test ./...

# 基准测试
go test -bench=. ./...

# 覆盖率
go test -cover ./...
```

### 代码检查
```bash
# 格式化
go fmt ./...

# 静态检查
go vet ./...

# Lint
golangci-lint run
```

## 📊 性能基准

**测试环境**: MacBook Pro M1, 网络延迟20ms

| 场景 | Python | Go | 提升 |
|------|--------|-----|------|
| 1000路径扫描 | 2.3s | 0.8s | **2.9x** |
| 5000路径扫描 | 11.5s | 3.2s | **3.6x** |
| 内存峰值 | 152MB | 73MB | **2.1x** |
| 启动时间 | 980ms | 65ms | **15x** |

## 🗺️ 路线图

- [x] v1.0.0 - 核心功能完整
  - [x] HTTP请求引擎
  - [x] 通配符检测
  - [x] 并发扫描
  - [x] 过滤器链
  - [x] 报告系统
- [x] v1.1.0 - 高级功能
  - [x] 递归扫描（3种模式）
  - [x] 会话管理
  - [x] 爬虫功能
  - [x] 暂停/恢复
- [x] v1.2.0 - 完善体验
  - [x] 内置字典
  - [x] 中文界面
  - [x] DNS缓存
- [ ] v1.3.0 - 性能优化
  - [ ] 连接池优化
  - [ ] 内存优化
  - [ ] pprof集成
- [ ] v1.4.0 - 扩展功能
  - [ ] SQLite报告器
  - [ ] Docker镜像
  - [ ] CI/CD

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📜 许可证

GPLv2 License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [dirsearch](https://github.com/maurosoria/dirsearch) - 原始Python版本
- 所有贡献者

## 📧 联系方式

- **Issues**: https://github.com/k3vi-07/dirsearch-go/issues
- **Discussions**: https://github.com/k3vi-07/dirsearch-go/discussions

---

**⭐ 如果这个项目对你有帮助，请给个Star！**
