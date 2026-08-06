package upload

import (
	"context"
	"mime/multipart"
	"testing"
)

func TestUploadSingleRejectsOversizedFileBeforeStorage(t *testing.T) {
	service := NewService(nil, 1, nil, nil)

	_, err := service.UploadSingle(context.Background(), &multipart.FileHeader{
		Filename: "too-large.txt",
		Size:     2 * 1024 * 1024,
	}, "uploads")
	if err == nil {
		t.Fatal("expected max file size error")
	}
}
