package s3

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/storage"
)

// Storage is an S3-compatible object storage implementation. It can connect
// to MinIO, RustFS, AWS S3, or another service that implements the S3 API.
type Storage struct {
	client    *minioClient.Client
	bucket    string
	publicURL string
}

func New(cfg config.S3Config) (storage.FileStorage, error) {
	lookup := minioClient.BucketLookupAuto
	if cfg.ForcePathStyle {
		lookup = minioClient.BucketLookupPath
	}

	client, err := minioClient.New(cfg.Endpoint, &minioClient.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: lookup,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check S3 bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minioClient.MakeBucketOptions{Region: cfg.Region}); err != nil {
			return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
		}
	}

	return &Storage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

func sanitizeS3Folder(folder string) (string, error) {
	raw := strings.ReplaceAll(folder, "\\", "/")
	for _, p := range strings.Split(raw, "/") {
		if p == ".." {
			return "", fmt.Errorf("invalid folder path")
		}
	}
	out := filepath.ToSlash(filepath.Clean(raw))
	return strings.Trim(out, "/"), nil
}

func (s *Storage) Upload(ctx context.Context, file *storage.FileHeader, folder string) (*storage.UploadResult, error) {
	folder, err := sanitizeS3Folder(folder)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(file.Filename)
	filename := idgen.GenerateStringID() + ext
	objectName := filename
	if folder != "" {
		objectName = folder + "/" + filename
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.client.PutObject(ctx, s.bucket, objectName, file.File, file.Size, minioClient.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	return &storage.UploadResult{
		Path:     objectName,
		URL:      s.GetURL(objectName),
		Size:     file.Size,
		MimeType: contentType,
	}, nil
}

func (s *Storage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minioClient.RemoveObjectOptions{})
}

func (s *Storage) GetURL(path string) string {
	if strings.TrimSpace(s.publicURL) == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.publicURL, "/"), s.bucket, strings.TrimLeft(path, "/"))
}

func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minioClient.StatObjectOptions{})
	if err != nil {
		errResp := minioClient.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.Code == "NoSuchObject" || errResp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
