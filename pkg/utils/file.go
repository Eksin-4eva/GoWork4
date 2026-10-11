package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// sniffBytes 是判断真实文件类型需要读取的字节数，与 net/http 的约定一致。
const sniffBytes = 512

// AllowedImageMIME 是允许上传的图片类型到扩展名的映射。
var AllowedImageMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// AllowedVideoMIME 是允许上传的视频类型到扩展名的映射。
var AllowedVideoMIME = map[string]string{
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",
}

// Upload 描述一个待上传的文件。
//
// 业务层只关心「内容 + 大小 + 原始文件名」，不关心它是从 multipart、
// 本地磁盘还是测试内存里来的。
type Upload struct {
	Reader   io.ReadSeeker
	Size     int64
	Filename string
}

// DetectContentType 通过读取文件头判断真实类型，并把读取位置复位。
//
// 只信客户端给的后缀是不够的：把 .php 改名成 .jpg 就能绕过。
func DetectContentType(f io.ReadSeeker) (string, error) {
	head := make([]byte, sniffBytes)
	n, err := f.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("utils.DetectContentType: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("utils.DetectContentType: seek back: %w", err)
	}
	return http.DetectContentType(head[:n]), nil
}

// PickExtension 依据嗅探出的类型与原始文件名选出扩展名。
//
// 嗅探优先（后缀可以被伪造）；嗅探不出结果时退回后缀白名单，
// 因为 Go 的类型嗅探表对 mov、mkv 这类容器格式覆盖不全。
func PickExtension(sniffed, filename string, allowed map[string]string) (string, error) {
	if ext, ok := allowed[sniffed]; ok {
		return ext, nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedExt := range allowed {
		if ext == allowedExt {
			return ext, nil
		}
	}
	return "", fmt.Errorf("不支持的文件类型: 嗅探结果 %q, 后缀 %q", sniffed, ext)
}
