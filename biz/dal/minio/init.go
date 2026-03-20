package minio

import (
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

func Init(cfg *Config) error {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("minio connect failed: %w", err)
	}

	Client = client

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("check bucket failed: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket failed: %w", err)
		}
	}

	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": "*"},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`
	err = client.SetBucketPolicy(ctx, cfg.Bucket, fmt.Sprintf(policy, cfg.Bucket))
	if err != nil {
		return fmt.Errorf("set bucket policy failed: %w", err)
	}

	return nil
}

func GetConfigFromEnv() *Config {
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	return &Config{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		UseSSL:    useSSL,
		Bucket:    os.Getenv("MINIO_BUCKET"),
	}
}
