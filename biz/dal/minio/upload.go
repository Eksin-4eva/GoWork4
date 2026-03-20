package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

func getBucket() string {
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "biligo"
	}
	return bucket
}

func getEndpoint() string {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	return endpoint
}

func UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("minio client not initialized")
	}

	bucket := getBucket()
	_, err := Client.PutObject(ctx, bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload to minio failed: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", getEndpoint(), bucket, objectName)
	return url, nil
}

func UploadMultipartFile(ctx context.Context, folder string, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s/%d%s", folder, time.Now().UnixNano(), ext)

	contentType := "application/octet-stream"
	switch ext {
	case ".mp4":
		contentType = "video/mp4"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	}

	return UploadFile(ctx, objectName, src, file.Size, contentType)
}

func UploadBytes(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	return UploadFile(ctx, objectName, bytes.NewReader(data), int64(len(data)), contentType)
}

func GetObjectURL(objectName string) string {
	return fmt.Sprintf("http://%s/%s/%s", getEndpoint(), getBucket(), objectName)
}
