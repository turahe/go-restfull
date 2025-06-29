package media

import (
	"context"
	"mime/multipart"
	"webapi/config"
	"webapi/internal/db/model"
	"webapi/internal/dto"
	internal_minio "webapi/pkg/minio"

	"github.com/minio/minio-go/v7"
)

func (m *mediaApp) UploadAvatar(ctx context.Context, file *multipart.FileHeader) (*model.Media, error) {
	fileContent, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileContent.Close()

	conf := config.GetConfig().Minio

	objectName := file.Filename
	bucketName := conf.BucketName
	contentType := file.Header.Get("Content-Type")

	minioClient := internal_minio.GetMinio()
	if _, err = minioClient.PutObject(context.Background(), bucketName, objectName, fileContent, file.Size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return nil, err
	}
	media, err := m.Repo.Media.CreateMedia(ctx, model.Media{
		Name:     objectName,
		FileName: objectName,
		Size:     file.Size,
		MimeType: contentType,
	})
	if err != nil {
		return nil, err
	}
	err = m.Repo.Media.AttachMedia(ctx, dto.MediaRelation{
		MediaID:      media.ID,
		MediableType: media.MimeType,
		MediableId:   media.ID,
		Group:        "default",
	})
	if err != nil {
		return nil, err
	}

	return media, nil
}
