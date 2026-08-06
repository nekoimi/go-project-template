package factory

import (
	"testing"

	"github.com/nekoimi/go-project-template/internal/config"
)

func TestNewRejectsUnsupportedDriver(t *testing.T) {
	_, err := New(config.StorageConfig{Driver: "unknown"})
	if err == nil {
		t.Fatal("expected unsupported driver error")
	}
}

func TestNewCreatesLocalStorage(t *testing.T) {
	got, err := New(config.StorageConfig{
		Driver:  "local",
		BaseURL: "http://localhost/uploads",
		Local: config.LocalConfig{
			UploadDir: t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got == nil {
		t.Fatal("expected local storage")
	}
}
