package local

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/storage"
)

type localStorage struct {
	uploadDir string
	baseURL   string
}

func New(cfg config.StorageConfig) storage.FileStorage {
	return &localStorage{
		uploadDir: cfg.Local.UploadDir,
		baseURL:   cfg.BaseURL,
	}
}

// safeJoin ensures the resulting path is within the upload directory.
func (s *localStorage) safeJoin(elem ...string) (string, error) {
	target := filepath.Join(elem...)
	rel, err := filepath.Rel(s.uploadDir, target)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path traversal detected")
	}
	return target, nil
}

func (s *localStorage) Upload(_ context.Context, file *storage.FileHeader, folder string) (*storage.UploadResult, error) {
	// Sanitize folder: clean + reject traversal
	folder = filepath.Clean(folder)
	if strings.Contains(folder, "..") {
		return nil, fmt.Errorf("invalid folder path")
	}

	ext := filepath.Ext(file.Filename)
	filename := idgen.GenerateStringID() + ext

	destDir, err := s.safeJoin(s.uploadDir, folder)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	destPath := filepath.Join(destDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	written, err := io.Copy(dst, file.File)
	if err != nil {
		_ = os.Remove(destPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	relPath := filepath.Join(folder, filename)
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
	fullPath, err := s.safeJoin(s.uploadDir, path)
	if err != nil {
		return err
	}
	return os.Remove(fullPath)
}

func (s *localStorage) GetURL(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.baseURL, "/"), path)
}

func (s *localStorage) Exists(_ context.Context, path string) (bool, error) {
	fullPath, err := s.safeJoin(s.uploadDir, path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
