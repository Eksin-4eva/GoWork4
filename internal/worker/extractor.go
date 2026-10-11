package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// maxStderrLen 是错误信息里保留的 stderr 长度。
// ffmpeg 的输出很长，整段塞进日志没有意义。
const maxStderrLen = 500

// Extractor 从视频里抽一帧当封面。
//
// 抽成接口有两个好处：测试不必依赖本机装没装 ffmpeg；
// 而且「ffmpeg 缺失」这条失败路径可以被测到。
type Extractor interface {
	Cover(ctx context.Context, videoPath, outputPath string) error
}

// FFmpeg 是调用命令行 ffmpeg 的实现。
type FFmpeg struct {
	// Binary 是 ffmpeg 可执行文件路径，留空则用 PATH 里的 ffmpeg。
	Binary string
}

// Cover 把视频第 0 帧抽成一张图片。
func (f FFmpeg) Cover(ctx context.Context, videoPath, outputPath string) error {
	binary := f.Binary
	if binary == "" {
		binary = "ffmpeg"
	}

	cmd := exec.CommandContext(ctx, binary,
		"-y",            // 覆盖已存在的输出文件
		"-i", videoPath, // 输入
		"-frames:v", "1", // 只要第一帧
		"-q:v", "2", // 输出质量（2 接近无损）
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf(
				"worker.FFmpeg: 未找到 ffmpeg（装到 PATH，或用 Kafka.FfmpegPath 指定路径）: %w", err)
		}
		return fmt.Errorf("worker.FFmpeg: 抽帧失败: %w, stderr: %s", err, tail(stderr.String()))
	}
	return nil
}

// tail 截取字符串末尾，用于限制日志长度。
func tail(s string) string {
	if len(s) <= maxStderrLen {
		return s
	}
	return s[len(s)-maxStderrLen:]
}
