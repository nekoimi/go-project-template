package local

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/storage"
)

type localStorage struct {
	uploadDir string
	baseURL   string
	maxSize   int64 // bytes
}

func New(cfg config.StorageConfig) storage.FileStorage {
	return &localStorage{
		uploadDir: cfg.Local.UploadDir,
		baseURL:   cfg.BaseURL,
		maxSize:   int64(cfg.Local.MaxFileSize) * 1024 * 1024,
	}
}

func (s *localStorage) Upload(_ context.Context, file *storage.FileHeader, folder string) (*storage.UploadResult, error) {
	if file.Size > s.maxSize {
		return nil, fmt.Errorf("file size %d exceeds max allowed %d", file.Size, s.maxSize)
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext

	// 目标目录
	destDir := filepath.Join(s.uploadDir, folder)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	destPath := filepath.Join(destDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file.File)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	relPath := filepath.Join(folder, filename)
	// 统一使用正斜杠
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	mimeType := mime.TypeByExtension(ext)

	return &storage.UploadResult{
		Path:     relPath,
		URL:      s.GetURL(relPath),
		Size:     written,
		MimeType: mimeType,
	}, nil
}

func (s *localStorage) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.uploadDir, path)
	return os.Remove(fullPath)
}

func (s *localStorage) GetURL(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.baseURL, "/"), path)
}

func (s *localStorage) Exists(_ context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.uploadDir, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
