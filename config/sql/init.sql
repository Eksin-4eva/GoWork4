-- GoBili 数据库结构定义。
--
-- 这是全项目唯一的建表来源：
--   * 不使用 GORM AutoMigrate，结构变更一律改本文件
--   * 本地开发由 docker-compose 挂载到 mysql 的 /docker-entrypoint-initdb.d 自动执行
--   * 手动执行：mysql -uroot -p gobili < config/sql/init.sql（目标库需已存在）
--   * 测试库：make test-db 用本文件重建 gobili_test
--
-- 本文件只描述表结构，不含 CREATE DATABASE / USE：
-- 库的创建交给 docker-compose 的 MYSQL_DATABASE 或调用方的命令行参数，
-- 这样同一个文件既能用于开发库也能用于测试库。
--
-- 约定：
--   * 主键统一 bigint，由雪花算法在应用层生成，不用数据库自增
--   * 时间列统一用 datetime（不用 timestamp，避开时区隐式转换与 2038 问题）
--   * 软删除用 deleted_at，与 GORM 的 gorm.DeletedAt 对应
--   * 不建外键约束：外键在软删除场景下会互相牵制，且不利于将来分库

-- 用户
CREATE TABLE IF NOT EXISTS `users`
(
    `id`            bigint       NOT NULL COMMENT '雪花 ID',
    `username`      varchar(64)  NOT NULL COMMENT '用户名',
    `password_hash` varchar(255) NOT NULL COMMENT 'bcrypt 哈希',
    `avatar_key`    varchar(255) NOT NULL DEFAULT '' COMMENT 'MinIO object key，空表示未设置',
    `mfa_secret`    varchar(64)  NOT NULL DEFAULT '' COMMENT 'TOTP 密钥',
    `mfa_enabled`   tinyint(1)   NOT NULL DEFAULT 0 COMMENT '是否已开启 MFA',
    `created_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`    datetime     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    -- 软删除后用户名仍被占用，这是有意的：避免删号后用户名被他人抢注
    UNIQUE KEY `uniq_username` (`username`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '用户';

-- 刷新令牌
CREATE TABLE IF NOT EXISTS `refresh_tokens`
(
    `id`         bigint       NOT NULL COMMENT '雪花 ID',
    `user_id`    bigint       NOT NULL COMMENT '用户 ID',
    `token`      varchar(512) NOT NULL COMMENT 'JWT 原文',
    `expires_at` datetime     NOT NULL COMMENT '过期时间',
    `revoked`    tinyint(1)   NOT NULL DEFAULT 0 COMMENT '是否已吊销（登出）',
    `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_token` (`token`),
    KEY `idx_user_id` (`user_id`),
    -- 清理任务按过期时间扫描
    KEY `idx_expires_at` (`expires_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '刷新令牌';

-- 视频
CREATE TABLE IF NOT EXISTS `videos`
(
    `id`            bigint       NOT NULL COMMENT '雪花 ID',
    `user_id`       bigint       NOT NULL COMMENT '投稿人',
    `video_key`     varchar(255) NOT NULL COMMENT 'MinIO object key',
    `cover_key`     varchar(255) NOT NULL DEFAULT '' COMMENT '封面 key，空表示待 worker 抽帧生成',
    `title`         varchar(128) NOT NULL COMMENT '标题',
    `description`   text         NOT NULL COMMENT '简介',
    `visit_count`   bigint       NOT NULL DEFAULT 0 COMMENT '播放数',
    `comment_count` bigint       NOT NULL DEFAULT 0 COMMENT '评论数',
    `created_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`    datetime     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_visit_count` (`visit_count`),
    -- 封面补偿任务需要扫 cover_key = ''
    KEY `idx_cover_key` (`cover_key`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '视频';

-- 评论
CREATE TABLE IF NOT EXISTS `comments`
(
    `id`          bigint   NOT NULL COMMENT '雪花 ID',
    `user_id`     bigint   NOT NULL COMMENT '评论人',
    `video_id`    bigint   NOT NULL COMMENT '所属视频',
    -- 注意：0 表示顶层评论。旧版把空字符串写进这一列，在 MySQL 严格模式下会直接报错
    `parent_id`   bigint   NOT NULL DEFAULT 0 COMMENT '父评论 ID，0 表示顶层',
    `content`     text     NOT NULL COMMENT '内容',
    `child_count` bigint   NOT NULL DEFAULT 0 COMMENT '子评论数',
    `created_at`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  datetime NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    -- 评论列表固定按 (video_id, parent_id) 过滤再按时间倒序
    KEY `idx_video_parent_created` (`video_id`, `parent_id`, `created_at`),
    KEY `idx_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '评论';

-- 聊天消息
CREATE TABLE IF NOT EXISTS `chat_messages`
(
    `id`           bigint      NOT NULL COMMENT '雪花 ID',
    `room_id`      varchar(64) NOT NULL DEFAULT '' COMMENT '会话 ID，由双方 ID 排序后拼接',
    `sender_id`    bigint      NOT NULL COMMENT '发送方',
    `receiver_id`  bigint      NOT NULL COMMENT '接收方',
    `message_type` varchar(32) NOT NULL DEFAULT 'text' COMMENT '消息类型',
    `content`      text        NOT NULL COMMENT '内容',
    `read_at`      datetime    NULL     DEFAULT NULL COMMENT '已读时间',
    `created_at`   datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   datetime    NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    -- 拉取会话历史 / 拉取某个用户收到的消息
    KEY `idx_room_created` (`room_id`, `created_at`),
    KEY `idx_receiver_created` (`receiver_id`, `created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT '聊天消息';
