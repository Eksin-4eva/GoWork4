package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gobili/internal/event"
	"gobili/pkg/constants"
	"gobili/pkg/kafka"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// RunCoverGenerator 消费投稿事件，为没有封面的视频抽帧生成封面。
func (w *Worker) RunCoverGenerator(ctx context.Context) {
	messages := w.mq.Consumer.Start(ctx, constants.TopicVideoUploaded, constants.UploadedConsumerNum)

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			w.handleUploaded(ctx, msg)
		case <-ctx.Done():
			return
		}
	}
}

// RunCoverCompensator 定期扫描没有封面的视频并重发事件。
//
// 封面事件一旦丢了，用户就永远看不到封面——这是唯一「丢了无法自愈」的链路，
// 所以必须有兜底扫描，而不能只依赖消息本身。
func (w *Worker) RunCoverCompensator(ctx context.Context) {
	ticker := time.NewTicker(constants.CoverCompensateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.compensateCovers(ctx)
		}
	}
}

// handleUploaded 处理一条投稿事件。
//
// 失败只记日志、不做进程内重试：offset 会自动提交，重投要等进程重启；
// 真正的兜底是 RunCoverCompensator——它按 cover_key 为空来扫，
// 天然覆盖「事件丢了」与「处理失败」两种情况。
func (w *Worker) handleUploaded(ctx context.Context, msg *kafka.Message) {
	var ev event.VideoUploaded
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		logger.Errorf("worker: 解析投稿事件失败，跳过: %v", err)
		return
	}
	if ev.VideoID <= 0 || ev.VideoKey == "" {
		return
	}

	// 先查一次是为了省掉无谓的下载与抽帧；
	// 真正的并发保护在 SetCoverKey 的 WHERE 条件里。
	found, video, err := w.db.Video.GetByID(ctx, ev.VideoID)
	if err != nil {
		logger.Errorf("worker: 查询视频失败 video=%d: %v", ev.VideoID, err)
		return
	}
	if !found {
		logger.Warnf("worker: 视频不存在，跳过 video=%d", ev.VideoID)
		return
	}
	if video.CoverKey != "" {
		return
	}

	coverKey, err := w.generateCover(ctx, ev)
	if err != nil {
		logger.Errorf("worker: 生成封面失败 video=%d: %v", ev.VideoID, err)
		return
	}

	updated, err := w.db.Video.SetCoverKey(ctx, ev.VideoID, coverKey)
	if err != nil {
		logger.Errorf("worker: 回写封面失败 video=%d: %v", ev.VideoID, err)
		return
	}
	if !updated {
		logger.Infof("worker: 视频 %d 已有封面，本次回写被幂等护栏跳过", ev.VideoID)
		return
	}
	logger.Infof("worker: 视频 %d 封面已生成: %s", ev.VideoID, coverKey)
}

// generateCover 下载视频 → 抽帧 → 上传封面，返回封面 key。
func (w *Worker) generateCover(ctx context.Context, ev event.VideoUploaded) (string, error) {
	workDir, err := os.MkdirTemp("", "gobili-cover-")
	if err != nil {
		return "", fmt.Errorf("worker.generateCover: 创建临时目录: %w", err)
	}
	// 抽帧用的临时文件必须清理，否则磁盘会被慢慢吃满
	defer func() { _ = os.RemoveAll(workDir) }()

	videoPath := filepath.Join(workDir, "video"+filepath.Ext(ev.VideoKey))
	if err := w.storage.Download(ctx, ev.VideoKey, videoPath); err != nil {
		return "", err
	}

	coverPath := filepath.Join(workDir, "cover.jpg")
	if err := w.extractor.Cover(ctx, videoPath, coverPath); err != nil {
		return "", err
	}

	file, err := os.Open(coverPath)
	if err != nil {
		return "", fmt.Errorf("worker.generateCover: 打开封面: %w", err)
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("worker.generateCover: 读取封面信息: %w", err)
	}

	return w.storage.Put(ctx, constants.StoragePrefixCover, ev.VideoID,
		utils.AllowedImageMIME,
		utils.Upload{Reader: file, Size: info.Size(), Filename: "cover.jpg"}, "封面")
}

// compensateCovers 扫描 cover_key 为空的视频并重发投稿事件。
func (w *Worker) compensateCovers(ctx context.Context) {
	videos, err := w.db.Video.ListMissingCover(ctx, constants.CoverCompensateBatch)
	if err != nil {
		logger.Errorf("worker: 扫描缺失封面的视频失败: %v", err)
		return
	}
	if len(videos) == 0 {
		return
	}

	logger.Infof("worker: 发现 %d 个视频缺封面，重发事件", len(videos))
	for _, video := range videos {
		payload := event.VideoUploaded{
			V:        constants.SchemaVersion,
			VideoID:  video.ID,
			VideoKey: video.VideoKey,
		}
		key := []byte(strconv.FormatInt(video.ID, 10))
		if err := w.mq.Producer.Send(ctx, constants.TopicVideoUploaded, key, payload); err != nil {
			logger.Warnf("worker: 重发封面事件失败 video=%d: %v", video.ID, err)
		}
	}
}
