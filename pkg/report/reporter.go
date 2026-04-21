package report

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/youruser/dirsearch-go/pkg/fuzzer"
)

// Reporter 报告器接口
type Reporter interface {
	// Add 添加结果
	Add(result *fuzzer.ScanResult) error

	// Close 关闭报告器
	Close() error
}

// ReportFormat 报告格式
type ReportFormat string

const (
	FormatPlain    ReportFormat = "plain"
	FormatJSON     ReportFormat = "json"
	FormatCSV      ReportFormat = "csv"
	FormatXML      ReportFormat = "xml"
	FormatHTML     ReportFormat = "html"
	FormatMarkdown ReportFormat = "markdown"
)

// ReporterFactory 报告器工厂
type ReporterFactory struct{}

// NewReporter 创建报告器
func (f *ReporterFactory) NewReporter(format ReportFormat, filepath string) (Reporter, error) {
	switch format {
	case FormatJSON:
		return NewJSONReporter(filepath)
	case FormatCSV:
		return NewCSVReporter(filepath)
	case FormatXML:
		return NewXMLReporter(filepath)
	case FormatHTML:
		return NewHTMLReporter(filepath)
	case FormatMarkdown:
		return NewMarkdownReporter(filepath)
	default:
		return NewPlainReporter(filepath)
	}
}

// PlainReporter 简单文本报告器
type PlainReporter struct {
	file   *os.File
	writer *csv.Writer
}

// NewPlainReporter 创建简单文本报告器
func NewPlainReporter(filepath string) (*PlainReporter, error) {
	if filepath == "" {
		return &PlainReporter{file: os.Stdout}, nil
	}

	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &PlainReporter{
		file:   file,
		writer: csv.NewWriter(file),
	}, nil
}

// Add 添加结果
func (r *PlainReporter) Add(result *fuzzer.ScanResult) error {
	if r.writer != nil {
		record := []string{
			time.Now().Format("2006-01-02 15:04:05"),
			result.URL,
			result.Path,
			fmt.Sprintf("%d", result.Status),
			fmt.Sprintf("%d", result.Size),
			result.ContentType,
			result.Redirect,
		}
		return r.writer.Write(record)
	}

	return nil
}

// Close 关闭报告器
func (r *PlainReporter) Close() error {
	if r.writer != nil {
		r.writer.Flush()
	}
	if r.file != nil && r.file != os.Stdout {
		return r.file.Close()
	}
	return nil
}

// JSONReporter JSON报告器
type JSONReporter struct {
	file   *os.File
	result []*fuzzer.ScanResult
	mu     *customMutex
}

// NewJSONReporter 创建JSON报告器
func NewJSONReporter(filepath string) (*JSONReporter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &JSONReporter{
		file:   file,
		result: make([]*fuzzer.ScanResult, 0),
		mu:     newCustomMutex(),
	}, nil
}

// Add 添加结果
func (r *JSONReporter) Add(result *fuzzer.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.result = append(r.result, result)
	return nil
}

// Close 关闭报告器
func (r *JSONReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	encoder := json.NewEncoder(r.file)
	encoder.SetIndent("", "  ")

	data := map[string]interface{}{
		"info": map[string]string{
			"time": time.Now().Format("2006-01-02 15:04:05"),
		},
		"results": r.result,
	}

	if err := encoder.Encode(data); err != nil {
		return err
	}

	// 确保数据写入磁盘
	if err := r.file.Sync(); err != nil {
		return err
	}

	return r.file.Close()
}

// CSVReporter CSV报告器
type CSVReporter struct {
	file   *os.File
	writer *csv.Writer
	mu     *customMutex
	header bool
}

// NewCSVReporter 创建CSV报告器
func NewCSVReporter(filepath string) (*CSVReporter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &CSVReporter{
		file:   file,
		writer: csv.NewWriter(file),
		mu:     newCustomMutex(),
		header: false,
	}, nil
}

// Add 添加结果
func (r *CSVReporter) Add(result *fuzzer.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 写入表头
	if !r.header {
		header := []string{"Time", "URL", "Path", "Status", "Size", "ContentType", "Redirect"}
		if err := r.writer.Write(header); err != nil {
			return err
		}
		r.header = true
	}

	// 写入数据
	record := []string{
		result.Time.Format("2006-01-02 15:04:05"),
		result.URL,
		result.Path,
		fmt.Sprintf("%d", result.Status),
		fmt.Sprintf("%d", result.Size),
		result.ContentType,
		result.Redirect,
	}

	return r.writer.Write(record)
}

// Close 关闭报告器
func (r *CSVReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.writer.Flush()
	return r.file.Close()
}

// XMLReporter XML报告器
type XMLReporter struct {
	file   *os.File
	result ScanResults
	mu     *customMutex
}

// ScanResults 扫描结果集合
type ScanResults struct {
	XMLName xml.Name          `xml:"results"`
	Time    string            `xml:"time"`
	Results []fuzzer.ScanResult `xml:"result"`
}

// NewXMLReporter 创建XML报告器
func NewXMLReporter(filepath string) (*XMLReporter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &XMLReporter{
		file:   file,
		result: ScanResults{Time: time.Now().Format("2006-01-02 15:04:05")},
		mu:     newCustomMutex(),
	}, nil
}

// Add 添加结果
func (r *XMLReporter) Add(result *fuzzer.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.result.Results = append(r.result.Results, *result)
	return nil
}

// Close 关闭报告器
func (r *XMLReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	encoder := xml.NewEncoder(r.file)
	encoder.Indent("", "  ")

	if err := encoder.Encode(r.result); err != nil {
		return err
	}

	return r.file.Close()
}

// HTMLReporter HTML报告器
type HTMLReporter struct {
	file   *os.File
	result []*fuzzer.ScanResult
	mu     *customMutex
}

// htmlTemplate HTML模板
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>dirsearch-go Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .status-200 { color: green; }
        .status-300 { color: orange; }
        .status-400 { color: red; }
        .status-500 { color: darkred; }
    </style>
</head>
<body>
    <h1>dirsearch-go Scan Report</h1>
    <p>Time: {{.Time}}</p>
    <p>Total: {{.Total}}</p>

    <table>
        <tr>
            <th>Time</th>
            <th>URL</th>
            <th>Path</th>
            <th>Status</th>
            <th>Size</th>
            <th>Type</th>
        </tr>
        {{range .Results}}
        <tr>
            <td>{{.Time.Format "2006-01-02 15:04:05"}}</td>
            <td>{{.URL}}</td>
            <td>{{.Path}}</td>
            <td class="status-{{div .Status 100}}00">{{.Status}}</td>
            <td>{{.Size}}</td>
            <td>{{.ContentType}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`

// NewHTMLReporter 创建HTML报告器
func NewHTMLReporter(filepath string) (*HTMLReporter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &HTMLReporter{
		file:   file,
		result: make([]*fuzzer.ScanResult, 0),
		mu:     newCustomMutex(),
	}, nil
}

// Add 添加结果
func (r *HTMLReporter) Add(result *fuzzer.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.result = append(r.result, result)
	return nil
}

// Close 关闭报告器
func (r *HTMLReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	data := struct {
		Time    time.Time
		Results []*fuzzer.ScanResult
		Total   int
	}{
		Time:    time.Now(),
		Results: r.result,
		Total:   len(r.result),
	}

	if err := tmpl.Execute(r.file, data); err != nil {
		return err
	}

	return r.file.Close()
}

// MarkdownReporter Markdown报告器
type MarkdownReporter struct {
	file   *os.File
	result []*fuzzer.ScanResult
	mu     *customMutex
}

// NewMarkdownReporter 创建Markdown报告器
func NewMarkdownReporter(filepath string) (*MarkdownReporter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &MarkdownReporter{
		file:   file,
		result: make([]*fuzzer.ScanResult, 0),
		mu:     newCustomMutex(),
	}, nil
}

// Add 添加结果
func (r *MarkdownReporter) Add(result *fuzzer.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.result = append(r.result, result)
	return nil
}

// Close 关闭报告器
func (r *MarkdownReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Fprintf(r.file, "# dirsearch-go Report\n\n")
	fmt.Fprintf(r.file, "**Time:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(r.file, "**Total:** %d\n\n", len(r.result))

	fmt.Fprintf(r.file, "| Time | URL | Path | Status | Size | Type |\n")
	fmt.Fprintf(r.file, "|------|-----|------|--------|------|------|\n")

	for _, res := range r.result {
		fmt.Fprintf(r.file, "| %s | %s | %s | %d | %d | %s |\n",
			res.Time.Format("2006-01-02 15:04:05"),
			res.URL,
			res.Path,
			res.Status,
			res.Size,
			res.ContentType,
		)
	}

	return r.file.Close()
}

// customMutex 自定义互斥锁（简单包装）
type customMutex struct {
	mu chan struct{}
}

func newCustomMutex() *customMutex {
	return &customMutex{mu: make(chan struct{}, 1)}
}

func (m *customMutex) Lock() {
	m.mu <- struct{}{}
}

func (m *customMutex) Unlock() {
	<-m.mu
}
