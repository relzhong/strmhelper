package models

import (
	"gorm.io/gorm"
)

type StrmTask struct {
	gorm.Model
	TaskID          string `gorm:"uniqueIndex;column:task_id"` // Matches 'id' in yaml
	Cron            string
	URL             string `gorm:"column:url"`
	PublicURL       string `gorm:"column:public_url"`
	Username        string
	Password        string
	Token           string
	SourceDir       string
	TargetDir       string
	FlattenMode     bool
	Subtitle        bool
	Image           bool
	NFO             bool
	Mode            string // OpenListURL, RawURL, OpenListPath
	Overwrite       bool
	SyncServer      bool
	SyncIgnore      string
	ProtectEnabled  bool
	ProtectThreshold int
	ProtectGrace    int
	OtherExt        string
	MaxWorkers      int
	MaxDownloaders  int
	WaitTime        int
	CheckModTime    bool `gorm:"default:false"`
	SyncBack        bool `gorm:"default:false"`
	IsRunning       bool `gorm:"default:false"`
	LastError       string
	NextRun         string `gorm:"-"`
}

type SyncedFile struct {
	ID         uint   `gorm:"primarykey"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	UpdatedAt  int64  `gorm:"autoUpdateTime"`
	TaskID     string `gorm:"uniqueIndex:idx_task_local;uniqueIndex:idx_task_remote"`
	RemotePath string `gorm:"uniqueIndex:idx_task_remote"`
	LocalPath  string `gorm:"uniqueIndex:idx_task_local"`
}

type PendingRemoteDelete struct {
	ID           uint   `gorm:"primarykey"`
	CreatedAt    int64  `gorm:"autoCreateTime"`
	UpdatedAt    int64  `gorm:"autoUpdateTime"`
	TaskID       string `gorm:"uniqueIndex:idx_task_pending"`
	RemotePath   string `gorm:"uniqueIndex:idx_task_pending"`
	MissingCount int
}

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Password string
}
