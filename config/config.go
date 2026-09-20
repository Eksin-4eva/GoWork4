package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	mu  sync.RWMutex
	cfg *Config
)

// Init 从指定路径加载配置，可重复调用（后一次覆盖前一次）。
func Init(path string) error {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("config.Init: read config file %q: %w", path, err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return fmt.Errorf("config.Init: unmarshal config: %w", err)
	}

	mu.Lock()
	cfg = &c
	mu.Unlock()
	return nil
}

// Get 返回当前配置。调用前必须先 Init，否则 panic。
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()

	if cfg == nil {
		panic("config.Get: config is not initialized, call config.Init first")
	}
	return cfg
}
