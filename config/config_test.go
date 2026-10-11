package config

import (
	"strings"
	"testing"
	"time"
)

func TestInitLoadsExampleConfig(t *testing.T) {
	if err := Init("config.example.yaml"); err != nil {
		t.Fatalf("Init: %v", err)
	}

	cfg := Get()

	if cfg.Server.Name != "gobili-api" {
		t.Errorf("Server.Name = %q", cfg.Server.Name)
	}
	if cfg.Server.Port != 8888 {
		t.Errorf("Server.Port = %d", cfg.Server.Port)
	}
	if cfg.MySQL.Database != "gobili" {
		t.Errorf("MySQL.Database = %q", cfg.MySQL.Database)
	}
	if cfg.MinIO.PublicBaseURL != "http://127.0.0.1:9000/gobili" {
		t.Errorf("MinIO.PublicBaseURL = %q", cfg.MinIO.PublicBaseURL)
	}
	// viper 需要把 "2h" 这类字符串解码成 time.Duration
	if cfg.JWT.AccessTTL != 2*time.Hour {
		t.Errorf("JWT.AccessTTL = %v, want 2h", cfg.JWT.AccessTTL)
	}
	if cfg.JWT.RefreshTTL != 168*time.Hour {
		t.Errorf("JWT.RefreshTTL = %v, want 168h", cfg.JWT.RefreshTTL)
	}
}

func TestInitFailsOnMissingFile(t *testing.T) {
	if err := Init("no-such-config.yaml"); err == nil {
		t.Fatal("Init should fail for a missing file")
	}
}

func TestMySQLDSN(t *testing.T) {
	m := MySQL{
		Host: "127.0.0.1", Port: 3306, Database: "gobili",
		Username: "gobili", Password: "gobili", Charset: "utf8mb4",
		Loc: "Asia/Shanghai", TimeZone: "+08:00",
	}
	want := "gobili:gobili@tcp(127.0.0.1:3306)/gobili?" +
		"charset=utf8mb4&loc=Asia%2FShanghai&parseTime=true&time_zone=%27%2B08%3A00%27"
	if got := m.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

// 会话时区必须下发到数据库，否则 created_at（应用写入）与
// updated_at（由 CURRENT_TIMESTAMP 生成）会用不同时区，相差 8 小时。
func TestMySQLDSNSetsSessionTimeZone(t *testing.T) {
	m := MySQL{
		Host: "127.0.0.1", Port: 3306, Database: "gobili",
		Username: "gobili", Password: "gobili", Charset: "utf8mb4",
		Loc: "Asia/Shanghai", TimeZone: "+08:00",
	}
	if !strings.Contains(m.DSN(), "time_zone=") {
		t.Fatalf("DSN 必须带上 time_zone: %s", m.DSN())
	}
}
