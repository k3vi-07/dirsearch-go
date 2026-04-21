package fuzzer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/youruser/dirsearch-go/pkg/dictionary"
	"github.com/youruser/dirsearch-go/pkg/requester"
	"github.com/youruser/dirsearch-go/pkg/scanner"
)

// Fuzzer 模糊测试器接口
type Fuzzer interface {
	// Start 开始扫描
	Start(ctx context.Context) error

	// Pause 暂停扫描
	Pause() error

	// Resume 恢复扫描
	Resume() error

	// Stop 停止扫描
	Stop() error

	// GetStats 获取统计信息
	GetStats() *Stats
}

// Stats 统计信息
type Stats struct {
	TotalScanned  int
	Found         int
	Errors        int
	StartTime     time.Time
	EndTime       time.Time
	RequestRate   float64
	mu            sync.RWMutex
}

// AsyncFuzzer 异步模糊测试器
type AsyncFuzzer struct {
	requester    requester.Requester
	dictionary   *dictionary.Dictionary
	scannerSet   *scanner.ScannerSet
	filterChain  *FilterChain
	workers      int
	resultChan   chan *ScanResult
	errorChan    chan *error
	basePath     string
	stats        *Stats
	playEvent    *playPauseEvent
	quitEvent    *quitEvent
	wg           sync.WaitGroup
}

// ScanResult 扫描结果
type ScanResult struct {
	URL      string
	Path     string
	Status   int
	Size     int64
	ContentType string
	Redirect string
	Time     time.Time
}

// FilterChain 过滤器链
type FilterChain struct {
	filters []scanner.Filter
}

// NewFilterChain 创建过滤器链
func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]scanner.Filter, 0),
	}
}

// AddFilter 添加过滤器
func (c *FilterChain) AddFilter(filter scanner.Filter) {
	c.filters = append(c.filters, filter)
}

// IsExcluded 检查是否被排除
func (c *FilterChain) IsExcluded(resp *requester.Response) bool {
	for _, filter := range c.filters {
		if filter.IsExcluded(resp) {
			return true
		}
	}
	return false
}

// NewAsyncFuzzer 创建异步模糊测试器
func NewAsyncFuzzer(
	req requester.Requester,
	dict *dictionary.Dictionary,
	scanners *scanner.ScannerSet,
	workers int,
	basePath string,
) *AsyncFuzzer {
	return &AsyncFuzzer{
		requester:   req,
		dictionary:  dict,
		scannerSet:  scanners,
		filterChain: NewFilterChain(),
		workers:     workers,
		resultChan:  make(chan *ScanResult, 100),
		errorChan:   make(chan *error, 10),
		basePath:    basePath,
		stats:       &Stats{StartTime: time.Now()},
		playEvent:   newPlayPauseEvent(),
		quitEvent:   newQuitEvent(),
	}
}

// Start 开始扫描
func (f *AsyncFuzzer) Start(ctx context.Context) error {
	f.stats.StartTime = time.Now()
	f.playEvent.play()

	// 初始化扫描器
	if err := f.setupScanners(ctx); err != nil {
		return fmt.Errorf("扫描器初始化失败: %w", err)
	}

	// 启动结果处理器
	go f.processResults(ctx)

	// 启动worker池
	for i := 0; i < f.workers; i++ {
		f.wg.Add(1)
		go f.worker(ctx, i)
	}

	// 等待所有worker完成
	f.wg.Wait()

	// 关闭通道
	close(f.resultChan)
	close(f.errorChan)

	f.stats.EndTime = time.Now()
	f.calculateRequestRate()

	return nil
}

// worker 工作协程
func (f *AsyncFuzzer) worker(ctx context.Context, id int) {
	defer f.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if f.quitEvent.isQuitSet() {
				return
			}
			// 等待播放信号
			if !f.playEvent.isPlaying() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 获取下一个路径
			path, ok := f.dictionary.Next()
			if !ok {
				return
			}

			// 执行扫描
			if err := f.scanPath(ctx, path); err != nil {
				f.stats.mu.Lock()
				f.stats.Errors++
				f.stats.mu.Unlock()

				select {
				case f.errorChan <- &err:
				default:
				}
			}
		}
	}
}

// scanPath 扫描单个路径
func (f *AsyncFuzzer) scanPath(ctx context.Context, path string) error {
	// 发送请求
	resp, err := f.requester.Request(ctx, path)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}

	f.stats.mu.Lock()
	f.stats.TotalScanned++
	f.stats.mu.Unlock()

	// 过滤响应
	if f.filterChain.IsExcluded(resp) {
		return nil
	}

	// 获取适用的扫描器
	wildcardScanner, ok := f.scannerSet.Get("default")
	if !ok || !wildcardScanner.IsWildcard() {
		// 没有通配符，直接添加结果
		f.addResult(resp)
		return nil
	}

	// 检查通配符
	if wildcardScanner.Check(resp) {
		f.addResult(resp)
	}

	return nil
}

// addResult 添加结果
func (f *AsyncFuzzer) addResult(resp *requester.Response) {
	f.stats.mu.Lock()
	f.stats.Found++
	f.stats.mu.Unlock()

	result := &ScanResult{
		URL:         resp.URL,
		Path:        resp.Path,
		Status:      resp.Status,
		Size:        resp.Length,
		ContentType: resp.ContentType,
		Redirect:    resp.Redirect,
		Time:        time.Now(),
	}

	select {
	case f.resultChan <- result:
	default:
		// 通道满，丢弃结果
	}
}

// setupScanners 初始化扫描器
func (f *AsyncFuzzer) setupScanners(ctx context.Context) error {
	// 创建默认扫描器
	wildcardScanner := scanner.NewWildcardScanner(f.requester, f.basePath)

	// 设置扫描器基础URL
	f.requester.SetURL(f.basePath)

	// 初始化通配符检测
	if err := wildcardScanner.Setup(ctx); err != nil {
		return err
	}

	f.scannerSet.Set("default", wildcardScanner)

	return nil
}

// processResults 处理结果
func (f *AsyncFuzzer) processResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-f.resultChan:
			if !ok {
				return
			}
			// TODO: 发送到报告器
			fmt.Printf("[+] %d - %s - %d\n", result.Status, result.Path, result.Size)
		case err, ok := <-f.errorChan:
			if !ok {
				return
			}
			// 记录错误
			fmt.Printf("[-] Error: %v\n", err)
		}
	}
}

// Pause 暂停扫描
func (f *AsyncFuzzer) Pause() error {
	f.playEvent.pause()
	return nil
}

// Resume 恢复扫描
func (f *AsyncFuzzer) Resume() error {
	f.playEvent.play()
	return nil
}

// Stop 停止扫描
func (f *AsyncFuzzer) Stop() error {
	f.quitEvent.setQuit()
	f.playEvent.play() // 确保worker不会卡住
	return nil
}

// GetStats 获取统计信息
func (f *AsyncFuzzer) GetStats() *Stats {
	f.calculateRequestRate()
	return f.stats
}

// calculateRequestRate 计算请求速率
func (f *AsyncFuzzer) calculateRequestRate() {
	f.stats.mu.Lock()
	defer f.stats.mu.Unlock()

	if f.stats.EndTime.IsZero() {
		f.stats.EndTime = time.Now()
	}

	duration := f.stats.EndTime.Sub(f.stats.StartTime).Seconds()
	if duration > 0 {
		f.stats.RequestRate = float64(f.stats.TotalScanned) / duration
	}
}

// playPauseEvent 播放/暂停事件
type playPauseEvent struct {
	playing bool
	mu      sync.RWMutex
}

func newPlayPauseEvent() *playPauseEvent {
	return &playPauseEvent{playing: false}
}

func (e *playPauseEvent) play() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playing = true
}

func (e *playPauseEvent) pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.playing = false
}

func (e *playPauseEvent) isPlaying() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.playing
}

// quitEvent 退出事件
type quitEvent struct {
	isQuit bool
	mu     sync.RWMutex
}

func newQuitEvent() *quitEvent {
	return &quitEvent{isQuit: false}
}

func (e *quitEvent) setQuit() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.isQuit = true
}

func (e *quitEvent) isQuitSet() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isQuit
}
