# 字典说明

本目录包含从[dirsearch](https://github.com/maurosoria/dirsearch)复制的完整字典集合。

## 主字典

- **common.txt** (9680行) - 默认主字典，包含所有常见Web路径
- **dicc.txt** (9680行) - 原始完整字典备份
- **common_small.txt** (446行) - 精简版字典（快速扫描用）

## 黑名单

- **400_blacklist.txt** - HTTP 400状态码黑名单
- **403_blacklist.txt** - HTTP 403状态码黑名单
- **500_blacklist.txt** - HTTP 500状态码黑名单

## 分类字典

### 按技术栈分类

#### PHP
- **php/wordpress.txt** - WordPress路径
- **php/drupal.txt** - Drupal路径
- **php/joomla.txt** - Joomla路径
- **php/magento.txt** - Magento路径
- **php/laravel.txt** - Laravel路径
- **php/symfony.txt** - Symfony路径
- **php/cakephp.txt** - CakePHP路径
- **php/codeigniter.txt** - CodeIgniter路径
- **php/yii.txt** - Yii路径
- **php/plugins-full.txt** - 完整插件列表
- **php/plugins-vulnerable.txt** - 已知漏洞插件

#### Java
- **java/jsp.txt** - JSP路径
- **java/jsf.txt** - JSF路径
- **java/spring.txt** - Spring框架路径

#### .NET
- **dotnet/aspx.txt** - ASPX路径
- **dotnet/core.txt** - .NET Core路径
- **dotnet/mvc.txt** - MVC路径

#### Python
- **python/django.txt** - Django路径
- **python/flask.txt** - Flask路径
- **python/fastapi.txt** - FastAPI路径

#### Node.js
- **node/express.txt** - Express框架路径

#### ColdFusion
- **coldfusion/coldfusion.txt** - ColdFusion路径

### 按功能分类

- **backups.txt** - 备份文件路径
- **conf.txt** - 配置文件路径
- **db.txt** - 数据库文件路径
- **extensions.txt** - 扩展名列表
- **keys.txt** - 密钥文件路径
- **logs.txt** - 日志文件路径
- **vcs.txt** - 版本控制路径（.git/.svn）
- **web.txt** - Web服务器路径

### 按基础设施分类

- **infra/aws.txt** - AWS相关路径
- **infra/docker.txt** - Docker相关路径
- **infra/k8s.txt** - Kubernetes相关路径

## 其他

- **user-agents.txt** - User-Agent字符串列表
- **test.txt** - 测试用小字典

## 使用方法

### 使用主字典
```bash
./dirsearch -u http://example.com
# 自动使用内置的common.txt (9680条)
```

### 使用分类字典
```bash
./dirsearch -u http://example.com -w db/categories/php/wordpress.txt
```

### 使用多个字典
```bash
./dirsearch -u http://example.com -w db/categories/common.txt,db/categories/conf.txt,db/categories/backups.txt
```

### 按技术栈扫描
```bash
# PHP应用
./dirsearch -u http://example.com -w db/categories/php/*.txt

# Java应用
./dirsearch -u http://example.com -w db/categories/java/*.txt

# 全栈扫描
./dirsearch -u http://example.com -w db/categories/*.txt
```

## 字典统计

| 类别 | 文件数 | 总路径数 |
|------|--------|----------|
| 主字典 | 3 | 16,844 |
| 黑名单 | 3 | - |
| PHP分类 | 12 | - |
| Java分类 | 3 | - |
| Python分类 | 3 | - |
| .NET分类 | 3 | - |
| 功能分类 | 8 | - |
| 基础设施分类 | 3 | - |
| 其他 | 2 | - |
| **总计** | **42** | **16,844+** |

## 许可证

这些字典文件来自[dirsearch](https://github.com/maurosoria/dirsearch)项目，遵循GPLv2许可证。
