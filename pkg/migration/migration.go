package migration

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DSN string
}

type Migrator struct {
	dsn       string
	dbName    string
	rootDSN   string
	tableDDLs []string
}

func NewMigrator(cfg Config) *Migrator {
	return &Migrator{
		dsn: cfg.DSN,
	}
}

func (m *Migrator) parseDSN() error {
	dsn := m.dsn
	var dsnWithoutDB string

	if strings.Contains(dsn, "tcp(") {
		parts := strings.SplitN(dsn, "/", 2)
		if len(parts) == 2 {
			dsnWithoutDB = parts[0] + "/"
			dbAndParams := parts[1]
			if idx := strings.Index(dbAndParams, "?"); idx != -1 {
				m.dbName = dbAndParams[:idx]
				params := dbAndParams[idx:]
				m.rootDSN = dsnWithoutDB + params
			} else {
				m.dbName = dbAndParams
				m.rootDSN = dsnWithoutDB
			}
		}
	} else {
		parts := strings.SplitN(dsn, "/", 2)
		if len(parts) == 2 {
			dsnWithoutDB = parts[0] + "/"
			dbAndParams := parts[1]
			if idx := strings.Index(dbAndParams, "?"); idx != -1 {
				m.dbName = dbAndParams[:idx]
				params := dbAndParams[idx:]
				m.rootDSN = dsnWithoutDB + params
			} else {
				m.dbName = dbAndParams
				m.rootDSN = dsnWithoutDB
			}
		}
	}

	if m.dbName == "" {
		return fmt.Errorf("cannot parse database name from DSN")
	}

	return nil
}

func (m *Migrator) createDatabase() error {
	db, err := sql.Open("mysql", m.rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql server: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping mysql server: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		m.dbName,
	))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func (m *Migrator) createTables() error {
	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	for _, ddl := range m.tableDDLs {
		if _, err := db.Exec(ddl); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (m *Migrator) Migrate() error {
	if err := m.parseDSN(); err != nil {
		return err
	}

	if err := m.createDatabase(); err != nil {
		return err
	}

	if err := m.createTables(); err != nil {
		return err
	}

	return nil
}

func GetGatewayTableDDLs() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(32) PRIMARY KEY,
			username VARCHAR(64) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			avatar_url VARCHAR(255) NOT NULL DEFAULT '',
			mfa_secret VARCHAR(64) NOT NULL DEFAULT '',
			mfa_enabled TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			token VARCHAR(512) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			revoked TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_refresh_tokens_user_id (user_id),
			CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS videos (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			video_url VARCHAR(255) NOT NULL,
			cover_url VARCHAR(255) NOT NULL,
			title VARCHAR(128) NOT NULL,
			description TEXT NOT NULL,
			visit_count BIGINT NOT NULL DEFAULT 0,
			like_count BIGINT NOT NULL DEFAULT 0,
			comment_count BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_videos_user_id (user_id),
			INDEX idx_videos_created_at (created_at),
			INDEX idx_videos_visit_count (visit_count),
			FULLTEXT KEY ft_videos_title_description (title, description),
			CONSTRAINT fk_videos_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS comments (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			video_id VARCHAR(32) NOT NULL,
			parent_id VARCHAR(32) NOT NULL DEFAULT '',
			like_count BIGINT NOT NULL DEFAULT 0,
			child_count BIGINT NOT NULL DEFAULT 0,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_comments_video_id (video_id),
			INDEX idx_comments_parent_id (parent_id),
			INDEX idx_comments_user_id (user_id),
			CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_comments_video FOREIGN KEY (video_id) REFERENCES videos(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS likes (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			video_id VARCHAR(32) NOT NULL DEFAULT '',
			comment_id VARCHAR(32) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			UNIQUE KEY uniq_user_video_like (user_id, video_id, comment_id),
			INDEX idx_likes_user_id (user_id),
			INDEX idx_likes_video_id (video_id),
			INDEX idx_likes_comment_id (comment_id),
			CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS relations (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			to_user_id VARCHAR(32) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			UNIQUE KEY uniq_relation (user_id, to_user_id),
			INDEX idx_relations_to_user_id (to_user_id),
			CONSTRAINT fk_relations_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_relations_to_user FOREIGN KEY (to_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS image_searches (
			id VARCHAR(32) PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL DEFAULT '',
			image_url VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_image_searches_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(32) PRIMARY KEY,
			from_user_id VARCHAR(32) NOT NULL,
			to_user_id VARCHAR(32) NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_messages_from_user_id (from_user_id),
			INDEX idx_messages_to_user_id (to_user_id),
			CONSTRAINT fk_messages_from_user FOREIGN KEY (from_user_id) REFERENCES users(id),
			CONSTRAINT fk_messages_to_user FOREIGN KEY (to_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS chat_messages (
			id VARCHAR(32) PRIMARY KEY,
			room_id VARCHAR(65) NOT NULL,
			sender_id VARCHAR(32) NOT NULL,
			receiver_id VARCHAR(32) NOT NULL,
			message_type VARCHAR(32) NOT NULL DEFAULT 'text',
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			INDEX idx_chat_messages_room_id (room_id),
			INDEX idx_chat_messages_sender_id (sender_id),
			INDEX idx_chat_messages_receiver_id (receiver_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
}

func RunMigration(dsn string) error {
	m := NewMigrator(Config{DSN: dsn})
	m.tableDDLs = GetGatewayTableDDLs()
	return m.Migrate()
}

func RunMigrationWithRetry(dsn string, maxRetries int, retryInterval time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := RunMigration(dsn); err != nil {
			lastErr = err
			time.Sleep(retryInterval)
			continue
		}
		return nil
	}
	return fmt.Errorf("migration failed after %d retries: %w", maxRetries, lastErr)
}
