package worker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"gobili/internal/event"
	"gobili/pkg/base/client"
	"gobili/pkg/constants"
	dbmodel "gobili/pkg/db/model"
	"gobili/pkg/kafka"
	"gobili/pkg/storage"
	"gobili/pkg/utils"
)

// stubExtractor 写一张最小 JPEG，替代真实的 ffmpeg。
//
// 抽帧本身是 ffmpeg 的职责；这里要验的是「下载 → 抽帧 → 上传 → 回写」这条链路，
// 以及抽帧失败时的行为。真实 ffmpeg 的调用另有单测覆盖。
type stubExtractor struct {
	err error
}

func (s stubExtractor) Cover(_ context.Context, _, outputPath string) error {
	if s.err != nil {
		return s.err
	}
	// 最小 JPEG：SOI + APP0/JFIF + EOI
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0xFF, 0xD9}
	return os.WriteFile(outputPath, data, 0o600)
}

// setupCoverWorker 在基础依赖之外再连上 MinIO。MinIO 不可用时跳过。
func setupCoverWorker(t *testing.T, extractor Extractor) (*Worker, *gorm.DB) {
	t.Helper()

	database, gormDB, c := setupDeps(t)

	minioClient, err := client.InitMinIO()
	if err != nil {
		t.Skipf("MinIO 不可用，跳过封面链路测试: %v", err)
	}
	return newWorker(database, c, storage.New(minioClient), nil, extractor), gormDB
}

// seedUploadedVideo 造一条「已上传视频、还没有封面」的记录，
// 并把假视频对象放进 MinIO，返回它真实的 object key。
func seedUploadedVideo(t *testing.T, w *Worker, gormDB *gorm.DB, id int64) string {
	t.Helper()

	if err := gormDB.Exec("TRUNCATE TABLE " + constants.VideoTableName).Error; err != nil {
		t.Fatalf("清空视频表失败: %v", err)
	}

	// 造一个能被嗅探成 video/mp4 的假文件
	path := filepath.Join(t.TempDir(), "video.mp4")
	content := append([]byte("\x00\x00\x00\x18ftypmp42"), make([]byte, 2048)...)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("造视频文件失败: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("打开视频文件失败: %v", err)
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("读取文件信息失败: %v", err)
	}

	key, err := w.storage.Put(context.Background(), constants.StoragePrefixVideo, id,
		utils.AllowedVideoMIME,
		utils.Upload{Reader: file, Size: info.Size(), Filename: "video.mp4"}, "视频")
	if err != nil {
		t.Fatalf("上传视频对象失败: %v", err)
	}

	video := &dbmodel.Video{ID: id, UserID: 1, VideoKey: key, Title: "t"}
	if err := gormDB.Table(constants.VideoTableName).Create(video).Error; err != nil {
		t.Fatalf("造视频记录失败: %v", err)
	}
	return key
}

func loadVideo(t *testing.T, gormDB *gorm.DB, id int64) dbmodel.Video {
	t.Helper()

	var video dbmodel.Video
	if err := gormDB.Table(constants.VideoTableName).Where("id = ?", id).First(&video).Error; err != nil {
		t.Fatalf("查询视频失败: %v", err)
	}
	return video
}

func TestHandleUploadedStoresCover(t *testing.T) {
	w, gormDB := setupCoverWorker(t, stubExtractor{})
	key := seedUploadedVideo(t, w, gormDB, testVideoID)

	payload, err := json.Marshal(event.VideoUploaded{
		V: constants.SchemaVersion, VideoID: testVideoID, VideoKey: key,
	})
	if err != nil {
		t.Fatalf("序列化事件失败: %v", err)
	}
	msg := &kafka.Message{Value: payload}

	w.handleUploaded(context.Background(), msg)

	video := loadVideo(t, gormDB, testVideoID)
	if video.CoverKey == "" {
		t.Fatal("封面 key 应当已被回写")
	}
	if !strings.HasPrefix(video.CoverKey, constants.StoragePrefixCover+"/") {
		t.Fatalf("封面 key 前缀不对: %s", video.CoverKey)
	}

	// Kafka 是 at-least-once，同一条事件必须能安全地重复处理
	w.handleUploaded(context.Background(), msg)
	if again := loadVideo(t, gormDB, testVideoID); again.CoverKey != video.CoverKey {
		t.Fatalf("重复处理不该改动已有封面：%s -> %s", video.CoverKey, again.CoverKey)
	}
}

// 抽帧失败时不能写入封面，也不能崩——封面留给补偿任务重试。
func TestHandleUploadedLeavesCoverEmptyOnExtractFailure(t *testing.T) {
	w, gormDB := setupCoverWorker(t, stubExtractor{err: errors.New("ffmpeg 未安装")})
	key := seedUploadedVideo(t, w, gormDB, testVideoID)

	payload, err := json.Marshal(event.VideoUploaded{
		V: constants.SchemaVersion, VideoID: testVideoID, VideoKey: key,
	})
	if err != nil {
		t.Fatalf("序列化事件失败: %v", err)
	}

	w.handleUploaded(context.Background(), &kafka.Message{Value: payload})

	if video := loadVideo(t, gormDB, testVideoID); video.CoverKey != "" {
		t.Fatalf("抽帧失败不该写入封面，实际 %q", video.CoverKey)
	}
}

// ffmpeg 不存在时要给出可读的错误，而不是让调用方去猜。
//
// 这台机器本来就没装 ffmpeg，所以这条用例同时是最真实的失败路径验证。
func TestFFmpegReportsMissingBinary(t *testing.T) {
	err := FFmpeg{Binary: "definitely-not-a-real-binary-9f8e7d"}.Cover(
		context.Background(), "in.mp4", "out.jpg")
	if err == nil {
		t.Fatal("ffmpeg 不存在时应当报错")
	}
	if !strings.Contains(err.Error(), "未找到 ffmpeg") {
		t.Fatalf("错误信息应当说明找不到 ffmpeg，实际: %v", err)
	}
}
