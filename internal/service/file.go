package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/nekoimi/go-project-template/internal/storage"
)

type FileService interface {
	UploadSingle(ctx context.Context, file *multipart.FileHeader, folder string) (*storage.UploadResult, error)
	UploadMultiple(ctx context.Context, files []*multipart.FileHeader, folder string) ([]*storage.UploadResult, error)
}

type fileService struct {
	storage storage.FileStorage
}

func NewFileService(storage storage.FileStorage) FileService {
	return &fileService{storage: storage}
}

func (s *fileService) UploadSingle(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (*storage.UploadResult, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fh := &storage.FileHeader{
		File:     file,
		Header:   fileHeader,
		Filename: fileHeader.Filename,
		Size:     fileHeader.Size,
	}

	return s.storage.Upload(ctx, fh, folder)
}

func (s *fileService) UploadMultiple(ctx context.Context, files []*multipart.FileHeader, folder string) ([]*storage.UploadResult, error) {
	var results []*storage.UploadResult

	for _, fileHeader := range files {
		ext := filepath.Ext(fileHeader.Filename)
		_ = ext // ext validation can be added here

		result, err := s.UploadSingle(ctx, fileHeader, folder)
		if err != nil {
			return results, fmt.Errorf("failed to upload %s: %w", fileHeader.Filename, err)
		}
		results = append(results, result)
	}

	return results, nil
}
