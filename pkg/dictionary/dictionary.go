package dictionary

import (
	_ "embed"
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/youruser/dirsearch-go/pkg/structures"
)

//go:embed common.txt
var defaultDictionaryBytes []byte

// Dictionary 字典管理器
type Dictionary struct {
	items      []string
	index      int
	extraItems []string // 额外发现的路径（递归/爬虫）
	extraIndex int
	mu         sync.Mutex
	extensions []string
	forceExt   bool
	lowercase  bool
	uppercase  bool
	capitalize bool
	prefixes   []string
	suffixes   []string
}

// NewDictionary 创建字典
func NewDictionary(files []string, extensions []string, forceExt bool) (*Dictionary, error) {
	dict := &Dictionary{
		extensions: extensions,
		forceExt:   forceExt,
		items:      make([]string, 0),
		extraItems: make([]string, 0),
	}

	if err := dict.loadFiles(files); err != nil {
		return nil, err
	}

	return dict, nil
}

// NewWithDefault 使用内置默认字典创建字典
func NewWithDefault(extensions []string, forceExt bool) (*Dictionary, error) {
	dict := &Dictionary{
		extensions: extensions,
		forceExt:   forceExt,
		items:      make([]string, 0),
		extraItems: make([]string, 0),
	}

	// 从嵌入的字节数组加载
	if err := dict.loadFromBytes(defaultDictionaryBytes); err != nil {
		return nil, err
	}

	return dict, nil
}

// loadFromBytes 从字节数组加载字典
func (d *Dictionary) loadFromBytes(data []byte) error {
	set := structures.NewOrderedSet()
	extRegex := regexp.MustCompile(`(?i)%ext%`)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "===") {
			continue
		}

		// 移除前导斜杠
		line = strings.TrimPrefix(line, "/")

		// 验证路径
		if !d.isValid(line) {
			continue
		}

		// 处理 %EXT% 标记
		if extRegex.MatchString(line) {
			for _, ext := range d.extensions {
				newline := extRegex.ReplaceAllString(line, ext)
				set.Add(newline)
			}
		} else {
			set.Add(line)

			// 强制扩展名（非黑名单）
			if d.forceExt && !strings.HasSuffix(line, "/") {
				// 添加目录版本
				set.Add(line + "/")

				// 添加带扩展名的版本
				if !strings.Contains(line, ".") {
					for _, ext := range d.extensions {
						set.Add(line + "." + ext)
					}
				}
			}
		}
	}

	// 应用大小写转换
	d.items = d.applyCaseConversion(set.ToList())

	return nil
}

// loadFiles 加载字典文件
func (d *Dictionary) loadFiles(files []string) error {
	set := structures.NewOrderedSet()
	extRegex := regexp.MustCompile(`(?i)%ext%`)

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("打开字典文件失败 %s: %w", file, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// 跳过空行和注释
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// 移除前导斜杠
			line = strings.TrimPrefix(line, "/")

			// 验证路径
			if !d.isValid(line) {
				continue
			}

			// 处理 %EXT% 标记
			if extRegex.MatchString(line) {
				for _, ext := range d.extensions {
					newline := extRegex.ReplaceAllString(line, ext)
					set.Add(newline)
				}
			} else {
				set.Add(line)

				// 强制扩展名（非黑名单）
				if d.forceExt && !strings.HasSuffix(line, "/") {
					// 添加目录版本
					set.Add(line + "/")

					// 添加带扩展名的版本
					if !strings.Contains(line, ".") {
						for _, ext := range d.extensions {
							set.Add(line + "." + ext)
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("读取字典文件失败 %s: %w", file, err)
		}
	}

	// 处理前缀和后缀
	if len(d.prefixes) > 0 || len(d.suffixes) > 0 {
		d.processPrefixesSuffixes(set)
	}

	// 应用大小写转换
	d.items = d.applyCaseConversion(set.ToList())

	return nil
}

// processPrefixesSuffixes 处理前缀和后缀
func (d *Dictionary) processPrefixesSuffixes(set *structures.OrderedSet) {
	processed := structures.NewOrderedSet()

	for _, path := range set.ToList() {
		// 添加原始路径
		processed.Add(path)

		// 添加前缀
		for _, prefix := range d.prefixes {
			if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, prefix) {
				processed.Add(prefix + path)
			}
		}

		// 添加后缀
		for _, suffix := range d.suffixes {
			if !strings.HasSuffix(path, "/") && !strings.HasSuffix(path, suffix) {
				processed.Add(path + suffix)
			}
		}
	}

	d.items = processed.ToList()
}

// applyCaseConversion 应用大小写转换
func (d *Dictionary) applyCaseConversion(items []string) []string {
	if d.lowercase {
		result := make([]string, len(items))
		for i, item := range items {
			result[i] = strings.ToLower(item)
		}
		return result
	}

	if d.uppercase {
		result := make([]string, len(items))
		for i, item := range items {
			result[i] = strings.ToUpper(item)
		}
		return result
	}

	if d.capitalize {
		result := make([]string, len(items))
		for i, item := range items {
			result[i] = strings.Title(item)
		}
		return result
	}

	return items
}

// isValid 验证路径
func (d *Dictionary) isValid(path string) bool {
	if path == "" || path == "#" {
		return false
	}

	// TODO: 添加排除扩展名检查

	return true
}

// Next 获取下一个路径
func (d *Dictionary) Next() (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 先处理主字典
	if d.index < len(d.items) {
		path := d.items[d.index]
		d.index++
		return path, true
	}

	// 再处理额外的路径（递归/爬虫发现）
	if d.extraIndex < len(d.extraItems) {
		path := d.extraItems[d.extraIndex]
		d.extraIndex++
		return path, true
	}

	return "", false
}

// Reset 重置索引
func (d *Dictionary) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.index = 0
}

// Length 获取字典长度
func (d *Dictionary) Length() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.items) + len(d.extraItems)
}

// Progress 获取进度
func (d *Dictionary) Progress() (current, total int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 计算当前索引
	currentIdx := d.index
	if currentIdx >= len(d.items) {
		// 已处理完主字典，计算额外路径的进度
		currentIdx = len(d.items) + d.extraIndex
	}

	total = len(d.items) + len(d.extraItems)
	current = currentIdx

	return
}

// SetPrefixes 设置前缀
func (d *Dictionary) SetPrefixes(prefixes []string) {
	d.prefixes = prefixes
}

// SetSuffixes 设置后缀
func (d *Dictionary) SetSuffixes(suffixes []string) {
	d.suffixes = suffixes
}

// SetLowercase 设置小写
func (d *Dictionary) SetLowercase(lowercase bool) {
	d.lowercase = lowercase
}

// SetUppercase 设置大写
func (d *Dictionary) SetUppercase(uppercase bool) {
	d.uppercase = uppercase
}

// SetCapitalization 设置首字母大写
func (d *Dictionary) SetCapitalization(capitalize bool) {
	d.capitalize = capitalize
}

// GetExtensions 获取扩展名
func (d *Dictionary) GetExtensions() []string {
	return d.extensions
}

// Clone 克隆字典（用于会话保存）
func (d *Dictionary) Clone() (*Dictionary, error) {
	cloned := &Dictionary{
		items:      make([]string, len(d.items)),
		index:      d.index,
		extensions: make([]string, len(d.extensions)),
		forceExt:   d.forceExt,
		lowercase:  d.lowercase,
		uppercase:  d.uppercase,
		capitalize: d.capitalize,
		prefixes:   make([]string, len(d.prefixes)),
		suffixes:   make([]string, len(d.suffixes)),
	}

	copy(cloned.items, d.items)
	copy(cloned.extensions, d.extensions)
	copy(cloned.prefixes, d.prefixes)
	copy(cloned.suffixes, d.suffixes)

	return cloned, nil
}

// GetState 获取状态（用于会话保存）
func (d *Dictionary) GetState() (items []string, index int, extraItems []string, extraIndex int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	items = make([]string, len(d.items))
	copy(items, d.items)

	extraItems = make([]string, len(d.extraItems))
	copy(extraItems, d.extraItems)

	return items, d.index, extraItems, d.extraIndex
}

// SetState 设置状态（用于会话恢复）
func (d *Dictionary) SetState(items []string, index int, extraItems []string, extraIndex int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.items = make([]string, len(items))
	copy(d.items, items)
	d.index = index

	d.extraItems = make([]string, len(extraItems))
	copy(d.extraItems, extraItems)
	d.extraIndex = extraIndex
}

// AddExtra 添加额外路径（递归/爬虫发现）
func (d *Dictionary) AddExtra(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 检查是否已存在
	for _, item := range d.extraItems {
		if item == path {
			return
		}
	}

	// 添加到额外路径列表
	d.extraItems = append(d.extraItems, path)
}

// GetExtraCount 获取额外路径数量
func (d *Dictionary) GetExtraCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.extraItems)
}

