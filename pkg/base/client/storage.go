package client

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"gobili/config"
	"gobili/pkg/constants"
	"gobili/pkg/logger"
)

// storageTimeout 是初始化时访问对象存储的超时。
const storageTimeout = 10 * time.Second

// InitMinIO 初始化 MinIO 客户端，并确保业务桶存在。
func InitMinIO() (*minio.Client, error) {
	c := config.Get().MinIO

	cli, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("client.InitMinIO: new client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), storageTimeout)
	defer cancel()

	exists, err := cli.BucketExists(ctx, constants.StorageBucket)
	if err != nil {
		return nil, fmt.Errorf("client.InitMinIO: check bucket: %w", err)
	}
	if !exists {
		if err := cli.MakeBucket(ctx, constants.StorageBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("client.InitMinIO: make bucket: %w", err)
		}
		logger.Infof("client.InitMinIO: bucket %q created", constants.StorageBucket)
	}
	return cli, nil
}
