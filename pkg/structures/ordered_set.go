package structures

import (
	"sync"
)

// OrderedSet 有序集合（保持插入顺序且去重）
type OrderedSet struct {
	items []string
	set   map[string]struct{}
	mu    sync.RWMutex
}

// NewOrderedSet 创建有序集合
func NewOrderedSet() *OrderedSet {
	return &OrderedSet{
		items: make([]string, 0),
		set:   make(map[string]struct{}),
	}
}

// Add 添加元素
func (s *OrderedSet) Add(item string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.set[item]; exists {
		return false
	}

	s.items = append(s.items, item)
	s.set[item] = struct{}{}
	return true
}

// AddAll 批量添加
func (s *OrderedSet) AddAll(items []string) {
	for _, item := range items {
		s.Add(item)
	}
}

// Contains 检查是否包含元素
func (s *OrderedSet) Contains(item string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.set[item]
	return exists
}

// ToList 转换为列表
func (s *OrderedSet) ToList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, len(s.items))
	copy(result, s.items)
	return result
}

// Size 获取大小
func (s *OrderedSet) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// IsEmpty 是否为空
func (s *OrderedSet) IsEmpty() bool {
	return s.Size() == 0
}

// Clear 清空集合
func (s *OrderedSet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make([]string, 0)
	s.set = make(map[string]struct{})
}

// Remove 移除元素
func (s *OrderedSet) Remove(item string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.set[item]; !exists {
		return false
	}

	delete(s.set, item)

	// 重建列表（保持顺序）
	newItems := make([]string, 0, len(s.items)-1)
	for _, i := range s.items {
		if i != item {
			newItems = append(newItems, i)
		}
	}
	s.items = newItems

	return true
}
