package config

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	AppVersion = "1.0.0" // Ideally read from elsewhere or injected via build
)

type Config struct {
	Settings struct {
		Dev bool `yaml:"DEV"`
	} `yaml:"Settings"`
	Web struct {
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
	} `yaml:"Web"`
	OpenList2StrmList []map[string]interface{} `yaml:"Alist2StrmList"`
}

type SettingManager struct {
	AppName    string
	AppVersion string
	TZ         string
	Debug      bool

	BaseDir   string
	ConfigDir string
	Config    Config
}

var Settings *SettingManager

func init() {
	Settings = &SettingManager{
		AppName:    "StrmHelper",
		AppVersion: AppVersion,
		TZ:         "Asia/Shanghai",
		Debug:      false,
	}
	Settings.initDirs()
	Settings.loadConfig()

	// Configure native slog logger (stdout only)
	level := slog.LevelInfo
	if Settings.Debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

func (s *SettingManager) initDirs() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	s.BaseDir = pwd
	s.ConfigDir = filepath.Join(s.BaseDir, "config")

	os.MkdirAll(s.ConfigDir, os.ModePerm)
}

func (s *SettingManager) loadConfig() {
	s.Config.Web.Username = "admin"
	s.Config.Web.Password = "admin"

	configPath := filepath.Join(s.ConfigDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &s.Config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	s.Debug = s.Config.Settings.Dev
}
