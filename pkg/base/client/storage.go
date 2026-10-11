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

	// 桶必须允许匿名读取，否则对外 URL 全部 403。
	// 这一步不能省：库里存的是 object key，客户端拿到的 URL 由
	// PublicBaseURL + key 拼出，桶保持私有的话前端拿到的头像和封面全是坏图。
	if err := cli.SetBucketPolicy(ctx, constants.StorageBucket, publicReadPolicy(constants.StorageBucket)); err != nil {
		return nil, fmt.Errorf("client.InitMinIO: set bucket policy: %w", err)
	}

	return cli, nil
}

// publicReadPolicy 生成「允许匿名读取桶内对象」的桶策略。
//
// 视频站的头像、封面、视频本身就是公开资源，不需要签名访问；
// 真正需要保护的写入口在应用侧（谁有权限上传），不在对象存储的读侧。
func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket)
}
