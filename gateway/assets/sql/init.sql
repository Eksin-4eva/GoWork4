CREATE DATABASE IF NOT EXISTS `Gobili` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `Gobili`;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(255) NOT NULL DEFAULT '',
  mfa_secret VARCHAR(64) NOT NULL DEFAULT '',
  mfa_enabled TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  token VARCHAR(512) NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  revoked TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL,
  INDEX idx_refresh_tokens_user_id (user_id),
  CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS videos (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
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
);

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  video_id BIGINT NOT NULL,
  parent_id BIGINT NOT NULL DEFAULT 0,
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
);

CREATE TABLE IF NOT EXISTS likes (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  video_id BIGINT NOT NULL DEFAULT 0,
  comment_id BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL,
  UNIQUE KEY uniq_user_video_like (user_id, video_id),
  UNIQUE KEY uniq_user_comment_like (user_id, comment_id),
  INDEX idx_likes_video_id (video_id),
  INDEX idx_likes_comment_id (comment_id),
  CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS relations (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL,
  UNIQUE KEY uniq_relation (user_id, to_user_id),
  INDEX idx_relations_to_user_id (to_user_id),
  CONSTRAINT fk_relations_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_relations_to_user FOREIGN KEY (to_user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS image_searches (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL DEFAULT 0,
  image_url VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL,
  INDEX idx_image_searches_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGINT PRIMARY KEY,
  room_id VARCHAR(64) NOT NULL DEFAULT '',
  sender_id BIGINT NOT NULL,
  receiver_id BIGINT NOT NULL DEFAULT 0,
  message_type VARCHAR(32) NOT NULL,
  content TEXT NOT NULL,
  read_at DATETIME NULL DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL DEFAULT NULL,
  INDEX idx_chat_messages_room_id (room_id),
  INDEX idx_chat_messages_sender_receiver (sender_id, receiver_id),
  CONSTRAINT fk_chat_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id)
);
