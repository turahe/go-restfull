package internal_minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"webapi/config"
	"webapi/internal/logger"
)

var (
	MinioClient *minio.Client
)

func Setup() error {
	var minioClient *minio.Client
	ctx := context.Background()

	conf := config.GetConfig()

	if conf.Minio.Enable {
		minioClient, err := minio.New(conf.Minio.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(conf.Minio.AccessKeyID, conf.Minio.AccessKeySecret, ""),
			Secure: conf.Minio.UseSSL,
		})
		if err != nil {
			logger.Log.Error("Failed to connect to Minio", zap.Error(err))
		}

		if _, err := minioClient.ListBuckets(ctx); err != nil {
			logger.Log.Error("Error connecting to Minio: %v", zap.Error(err))
			return err
		}
	}

	MinioClient = minioClient

	return nil
}

func IsAlive() bool {
	ctx := context.Background()

	if _, err := MinioClient.ListBuckets(ctx); err != nil {
		return false
	}

	return true
}

func GetMinio() *minio.Client {
	return MinioClient
}
