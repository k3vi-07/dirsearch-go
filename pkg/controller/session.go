package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/youruser/dirsearch-go/pkg/fuzzer"
)

const (
	// SessionVersion 会话格式版本
	SessionVersion = 1
)

// Session 会话数据
type Session struct {
	Version   int             `json:"version"`
	Metadata  SessionMetadata `json:"metadata"`
	Targets   []*TargetSession `json:"targets"`
	Dictionary *DictionarySession `json:"dictionary"`
}

// SessionMetadata 会话元数据
type SessionMetadata struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Command   string    `json:"command"`
	Version   string    `json:"version"`
}

// TargetSession 目标会话
type TargetSession struct {
	URL          string         `json:"url"`
	BasePath     string         `json:"base_path"`
	Results      []ScanResult   `json:"results"`
	Progress     int            `json:"progress"`
	Total        int            `json:"total"`
	Directories  []string       `json:"directories"`
	CurrentIndex  int            `json:"current_index"`
}

// ScanResult 扫描结果（序列化用）
type ScanResult struct {
	URL         string    `json:"url"`
	Path        string    `json:"path"`
	Status      int       `json:"status"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	Redirect    string    `json:"redirect"`
	Timestamp   time.Time `json:"timestamp"`
}

// DictionarySession 字典会话
type DictionarySession struct {
	Items      []string `json:"items"`
	Index      int      `json:"index"`
	ExtraItems []string `json:"extra_items"`
	ExtraIndex int      `json:"extra_index"`
	Extensions []string `json:"extensions"`
	ForceExt   bool     `json:"force_extensions"`
}

// SessionManager 会话管理器
type SessionManager struct {
	sessionDir string
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	homeDir, _ := os.UserHomeDir()
	sessionDir := filepath.Join(homeDir, ".dirsearch", "sessions")

	return &SessionManager{
		sessionDir: sessionDir,
	}
}

// Save 保存会话
func (sm *SessionManager) Save(sessionpath string, ctrl *Controller) error {
	// 确保目录存在
	dir := filepath.Dir(sessionpath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建会话目录失败: %w", err)
		}
	}

	// 创建会话数据
	session := &Session{
		Version: SessionVersion,
		Metadata: SessionMetadata{
			StartTime: time.Now(),
			EndTime:   time.Now(), // 保存时更新
			Command:   "dirsearch", // TODO: 获取完整命令
			Version:   "dev",
		},
		Dictionary: sm.createDictionarySession(ctrl),
		Targets:   []*TargetSession{
			sm.createTargetSession(ctrl),
		},
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化会话失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(sessionpath, data, 0644); err != nil {
		return fmt.Errorf("写入会话文件失败: %w", err)
	}

	return nil
}

// Load 加载会话
func (sm *SessionManager) Load(sessionpath string) (*Session, error) {
	// 读取文件
	data, err := os.ReadFile(sessionpath)
	if err != nil {
		return nil, fmt.Errorf("读取会话文件失败: %w", err)
	}

	// 解析 JSON
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("解析会话文件失败: %w", err)
	}

	// 验证版本
	if session.Version != SessionVersion {
		return nil, fmt.Errorf("不支持的会话版本: %d (当前版本: %d)", session.Version, SessionVersion)
	}

	return &session, nil
}

// List 列举会话
func (sm *SessionManager) List() ([]string, error) {
	if _, err := os.Stat(sm.sessionDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	files, err := os.ReadDir(sm.sessionDir)
	if err != nil {
		return nil, err
	}

	sessions := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			sessions = append(sessions, filepath.Join(sm.sessionDir, file.Name()))
		}
	}

	return sessions, nil
}

// Delete 删除会话
func (sm *SessionManager) Delete(sessionpath string) error {
	return os.Remove(sessionpath)
}

// createDictionarySession 创建字典会话
func (sm *SessionManager) createDictionarySession(ctrl *Controller) *DictionarySession {
	items, index, extraItems, extraIndex := ctrl.dictionary.GetState()

	return &DictionarySession{
		Items:      items,
		Index:      index,
		ExtraItems: extraItems,
		ExtraIndex: extraIndex,
		Extensions: ctrl.dictionary.GetExtensions(),
		ForceExt:   ctrl.config.ForceExtensions,
	}
}

// createTargetSession 创建目标会话
func (sm *SessionManager) createTargetSession(ctrl *Controller) *TargetSession {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()

	current, total := ctrl.dictionary.Progress()

	// 转换结果格式
	results := make([]ScanResult, len(ctrl.results))
	for i, r := range ctrl.results {
		results[i] = ScanResult{
			URL:         r.URL,
			Path:        r.Path,
			Status:      r.Status,
			Size:        r.Size,
			ContentType: r.ContentType,
			Redirect:    r.Redirect,
			Timestamp:   r.Time,
		}
	}

	return &TargetSession{
		URL:          ctrl.currentURL,
		BasePath:     ctrl.basePath,
		Results:      results,
		Progress:     current,
		Total:        total,
		Directories:  ctrl.directories,
		CurrentIndex: 0,
	}
}

// RestoreFromSession 从会话恢复
func (ctrl *Controller) RestoreFromSession(session *Session) error {
	// 恢复字典状态
	if session.Dictionary != nil {
		ctrl.dictionary.SetState(
			session.Dictionary.Items,
			session.Dictionary.Index,
			session.Dictionary.ExtraItems,
			session.Dictionary.ExtraIndex,
		)
	}

	// 恢复目标状态
	if len(session.Targets) > 0 {
		target := session.Targets[0]
		ctrl.currentURL = target.URL
		ctrl.basePath = target.BasePath
		ctrl.directories = target.Directories

		// 恢复结果
		ctrl.results = make([]*fuzzer.ScanResult, len(target.Results))
		for i, r := range target.Results {
			ctrl.results[i] = &fuzzer.ScanResult{
				URL:         r.URL,
				Path:        r.Path,
				Status:      r.Status,
				Size:        r.Size,
				ContentType: r.ContentType,
				Redirect:    r.Redirect,
				Time:        r.Timestamp,
			}
		}
	}

	return nil
}

// SaveSession 保存当前会话
func (ctrl *Controller) SaveSession(sessionpath string) error {
	sm := NewSessionManager()
	return sm.Save(sessionpath, ctrl)
}

// GetSessionInfo 获取会话信息
func (sm *SessionManager) GetSessionInfo(sessionpath string) (*SessionInfo, error) {
	session, err := sm.Load(sessionpath)
	if err != nil {
		return nil, err
	}

	info := &SessionInfo{
		Filepath:  sessionpath,
		URL:       session.Targets[0].URL,
		Progress:  float64(session.Targets[0].Progress) / float64(session.Targets[0].Total) * 100,
		Total:     session.Targets[0].Total,
		Scanned:   session.Targets[0].Progress,
		Results:   len(session.Targets[0].Results),
		StartTime: session.Metadata.StartTime,
		EndTime:   session.Metadata.EndTime,
	}

	return info, nil
}

// SessionInfo 会话信息
type SessionInfo struct {
	Filepath  string    `json:"filepath"`
	URL       string    `json:"url"`
	Progress float64   `json:"progress"`
	Total     int       `json:"total"`
	Scanned   int       `json:"scanned"`
	Results   int       `json:"results"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// FormatProgress 格式化进度
func (si *SessionInfo) FormatProgress() string {
	return fmt.Sprintf("%.2f%% (%d/%d)", si.Progress, si.Scanned, si.Total)
}
