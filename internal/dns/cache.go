package dns

import (
	"context"
	"net"
	"sync"
	"time"
)

// Cache DNS缓存
type Cache struct {
	mu    sync.RWMutex
	data  map[string][]net.IP
	cache map[string]time.Time
	ttl   time.Duration
}

// NewCache 创建DNS缓存
func NewCache() *Cache {
	return &Cache{
		data:  make(map[string][]net.IP),
		cache: make(map[string]time.Time),
		ttl:   5 * time.Minute, // 默认5分钟TTL
	}
}

// Get 获取缓存的DNS记录
func (c *Cache) Get(host string) ([]net.IP, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ips, exists := c.data[host]
	if !exists {
		return nil, false
	}

	// 检查TTL
	if cached, ok := c.cache[host]; ok {
		if time.Since(cached) > c.ttl {
			// 过期，删除缓存
			delete(c.data, host)
			delete(c.cache, host)
			return nil, false
		}
	}

	return ips, true
}

// Set 设置DNS缓存
func (c *Cache) Set(host string, ips []net.IP) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[host] = ips
	c.cache[host] = time.Now()
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string][]net.IP)
	c.cache = make(map[string]time.Time)
}

// Resolver DNS解析器
type Resolver struct {
	Cache     *Cache
	Interface string
}

// DialContext 自定义拨号函数
func (r *Resolver) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	// 检查缓存
	if ips, ok := r.Cache.Get(host); ok {
		// 使用缓存的IP
		return net.DialTimeout(network, net.JoinHostPort(ips[0].String(), port), 10*time.Second)
	}

	// 解析DNS
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	r.Cache.Set(host, ips)

	// 使用第一个IP
	return net.DialTimeout(network, net.JoinHostPort(ips[0].String(), port), 10*time.Second)
}

// Dial 自定义拨号函数（不带context）
func (r *Resolver) Dial(network, addr string) (net.Conn, error) {
	return r.DialContext(context.Background(), network, addr)
}
