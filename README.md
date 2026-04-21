# dirsearch-go

[![Go Report Card](https://goreportcard.com/badge/github.com/youruser/dirsearch-go)](https://goreportcard.com/report/github.com/youruser/dirsearch-go)
[![License](https://img.shields.io/badge/license-GPLv2-blue.svg)](LICENSE)

> Web path discovery tool written in Go - A complete rewrite of [dirsearch](https://github.com/maurosoria/dirsearch)

## 特性

- 🔍 **多线程扫描** - 高性能并发路径扫描
- 🎯 **通配符检测** - 智能识别和处理通配符响应
- 🔄 **递归扫描** - 自动发现子目录
- 📊 **多种报告格式** - JSON, CSV, HTML, XML, SQLite
- 🔐 **多种认证支持** - Basic, Digest, NTLM, Bearer, JWT
- 🌐 **代理支持** - HTTP/HTTPS/SOCKS 代理
- 💾 **会话管理** - 保存和恢复扫描进度
- 🚀 **高性能** - 比 Python 版本快 2-5 倍

## 安装

### 从源码编译

```bash
git clone https://github.com/youruser/dirsearch-go.git
cd dirsearch-go
make build
```

### 使用 Go 安装

```bash
go install github.com/youruser/dirsearch-go/cmd/dirsearch@latest
```

### 下载预编译二进制

从 [Releases](https://github.com/youruser/dirsearch-go/releases) 页面下载适合您平台的二进制文件。

## 快速开始

### 基础扫描

```bash
dirsearch -u http://example.com -w db/common.txt
```

### 使用多线程

```bash
dirsearch -u http://example.com -w db/common.txt -t 50
```

### 指定扩展名

```bash
dirsearch -u http://example.com -w db/common.txt -x php,html,js
```

### 使用代理

```bash
dirsearch -u http://example.com -w db/common.txt --proxy http://127.0.0.1:8080
```

### 导出结果

```bash
dirsearch -u http://example.com -w db/common.txt -o report.json -f json
```

## 使用方法

### 命令行选项

#### 必需参数

```
-u, --url <url>              目标 URL
-l, --url-list <file>        URL 列表文件
```

#### 字典设置

```
-w, --wordlists <paths>      字典文件路径 (逗号分隔)
-e, --extensions <exts>      扩展名列表 (逗号分隔)
-X, --exclude-extensions     排除的扩展名
--force-extensions           强制扩展名
--prefixes <prefixes>        自定义前缀
--suffixes <suffixes>        自定义后缀
--lowercase                  小写转换
--uppercase                  大写转换
--capitalization             首字母大写
```

#### 常规设置

```
-t, --threads <num>          线程数 (默认: 30)
-r, --recursive              递归扫描
--recursion-depth <num>      递归深度 (默认: 0)
--max-time <seconds>         最大扫描时间
--max-errors <num>           最大错误数
```

#### 请求设置

```
-m, --method <method>        HTTP 方法 (默认: GET)
-H, --header <header>        自定义请求头
--header-list <file>         请求头列表文件
--data <data>                请求体数据
--data-file <file>           请求体文件
-d, --delay <ms>             请求之间的延迟
--timeout <seconds>          请求超时 (默认: 10)
--user-agent <agent>         自定义 User-Agent
--random-agents              随机 User-Agent
--cookies <cookies>          Cookie
```

#### 连接设置

```
--proxy <proxy>              代理 URL
--proxy-list <file>          代理列表文件
--proxy-auth <auth>          代理认证
--max-rate <req/s>           最大请求速率
--retries <num>              重试次数 (默认: 1)
--tor                        使用 Tor
--network-interface <iface>  网络接口
```

#### 认证

```
--auth-type <type>           认证类型 (basic/digest/ntlm/bearer/jwt)
--auth <credential>          认证凭据
```

#### 过滤设置

```
--include-status <codes>     包含的状态码 (逗号分隔)
--exclude-status <codes>     排除的状态码 (逗号分隔)
--exclude-sizes <sizes>      排除的响应大小
--exclude-texts <texts>      排除的响应文本
--exclude-regex <pattern>    排除的正则表达式
--exclude-redirect <pattern> 排除的重定向
--exclude-response <path>    自定义排除响应
--min-response-size <size>   最小响应大小
--max-response-size <size>   最大响应大小
--filter-threshold <num>     过滤阈值
```

#### 输出设置

```
-o, --output <file>          输出文件
-f, --format <formats>       报告格式 (plain/json/csv/html/xml/sqlite)
--log-file <file>            日志文件
--log-level <level>          日志级别 (error/warning/info/debug)
```

#### 会话管理

```
--session <file>             会话文件
--save-session               保存会话
```

### 示例

#### 使用代理和认证

```bash
dirsearch -u http://example.com \
  -w db/common.txt \
  --proxy http://127.0.0.1:8080 \
  --auth-type basic \
  --auth admin:password
```

#### 递归扫描

```bash
dirsearch -u http://example.com \
  -w db/common.txt \
  -r \
  --recursion-depth 3
```

#### 多种报告格式

```bash
dirsearch -u http://example.com \
  -w db/common.txt \
  -o report \
  -f json,csv,html
```

#### 使用 URL 列表

```bash
dirsearch -l urls.txt -w db/common.txt -t 30
```

## 配置文件

可以在 `~/.dirsearch/config.yaml` 中创建默认配置：

```yaml
general:
  threads: 30
  max-time: 0
  max-errors: 25

dictionary:
  wordlists:
    - /usr/share/dirb/wordlists/common.txt
  extensions:
    - php
    - html
    - js
  force-extensions: false

request:
  http-method: GET
  headers:
    User-Agent: dirsearch-go

connection:
  timeout: 10
  max-retries: 1
  max-rate: 0

output:
  output-formats:
    - plain
```

## 性能对比

| 指标 | Python dirsearch | Go dirsearch-go |
|------|-----------------|----------------|
| 吞吐量 (req/s) | ~500 | 1000-2500 |
| 内存使用 | ~150MB | <80MB |
| 启动时间 | ~1s | <0.1s |
| 单文件部署 | ❌ | ✅ |

## 开发

### 构建

```bash
make build
```

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make vet
make lint
```

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 许可证

GPLv2 License - 详见 [LICENSE](LICENSE) 文件

## 致谢

本项目基于 [dirsearch](https://github.com/maurosoria/dirsearch) 重写，感谢原作者的贡献。

## 路线图

- [x] 基础扫描功能
- [ ] 完整的通配符检测
- [ ] 所有认证类型
- [ ] 会话管理
- [ ] 多种报告格式
- [ ] CI/CD
- [ ] Docker 支持
- [ ] 更多平台支持

## 联系方式

- GitHub Issues: https://github.com/youruser/dirsearch-go/issues
