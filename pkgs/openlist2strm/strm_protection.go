package openlist2strm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"log/slog"
)

type StrmProtectionManager struct {
	TargetDir  string
	StateFile  string
	Threshold  int
	GraceScans int
	Protected  map[string]int
}

func NewStrmProtectionManager(targetDir string, taskID string, threshold int, graceScans int) *StrmProtectionManager {
	m := &StrmProtectionManager{
		TargetDir:  targetDir,
		StateFile:  filepath.Join(targetDir, fmt.Sprintf(".strmhelper_strm_%s.json", taskID)),
		Threshold:  threshold,
		GraceScans: graceScans,
		Protected:  make(map[string]int),
	}
	m.Load()
	return m
}

func (m *StrmProtectionManager) ToRelative(filePath string) string {
	rel, err := filepath.Rel(m.TargetDir, filePath)
	if err != nil {
		return filePath
	}
	return rel
}

func (m *StrmProtectionManager) ToAbsolute(relPath string) string {
	return filepath.Join(m.TargetDir, relPath)
}

func (m *StrmProtectionManager) Load() {
	if _, err := os.Stat(m.StateFile); err == nil {
		data, err := os.ReadFile(m.StateFile)
		if err == nil {
			var state struct {
				Protected map[string]int `json:"protected"`
			}
			if err := json.Unmarshal(data, &state); err == nil && state.Protected != nil {
				m.Protected = state.Protected
			} else {
				slog.Warn("加载保护状态失败，重新开始")
			}
		}
	}
}

func (m *StrmProtectionManager) Save() {
	tempFile := m.StateFile + ".tmp"
	state := map[string]interface{}{
		"updated":   time.Now().Format(time.RFC3339),
		"protected": m.Protected,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		slog.Error(fmt.Sprintf("保护状态保存失败：%v", err))
		return
	}

	err = os.WriteFile(tempFile, data, 0644)
	if err != nil {
		slog.Error(fmt.Sprintf("保护状态保存失败：%v", err))
		return
	}

	os.Rename(tempFile, m.StateFile)
}

func (m *StrmProtectionManager) Process(strmToDelete map[string]bool, strmPresent map[string]bool) map[string]bool {
	returned := 0
	for relPath := range m.Protected {
		absPath := m.ToAbsolute(relPath)
		if strmPresent[absPath] {
			delete(m.Protected, relPath)
			returned++
		}
	}

	if returned > 0 {
		slog.Info(fmt.Sprintf("%d 个 .strm 文件已恢复，取消保护", returned))
	}

	if len(strmToDelete) < m.Threshold {
		if len(strmToDelete) > 0 {
			slog.Info(fmt.Sprintf("正常删除 %d 个 .strm（阈值：%d）", len(strmToDelete), m.Threshold))
		}
		return strmToDelete
	}

	slog.Warn(fmt.Sprintf("保护激活：%d 个 .strm 待删除（阈值：%d）", len(strmToDelete), m.Threshold))

	for filePath := range strmToDelete {
		relPath := m.ToRelative(filePath)
		m.Protected[relPath]++
	}

	readyRel := make(map[string]bool)
	for relPath, count := range m.Protected {
		if count >= m.GraceScans {
			readyRel[relPath] = true
		}
	}

	pending := len(m.Protected) - len(readyRel)
	ready := make(map[string]bool)

	if len(readyRel) > 0 {
		slog.Warn(fmt.Sprintf("删除 %d 个 .strm（经过 %d 次扫描确认）", len(readyRel), m.GraceScans))
		for relPath := range readyRel {
			ready[m.ToAbsolute(relPath)] = true
			delete(m.Protected, relPath)
		}
	}

	if pending > 0 {
		slog.Info(fmt.Sprintf("%d 个文件等待确认", pending))
	}

	return ready
}
