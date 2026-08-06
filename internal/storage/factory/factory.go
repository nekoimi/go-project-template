package factory

import (
	"fmt"
	"strings"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/storage"
	"github.com/nekoimi/go-project-template/internal/storage/local"
	"github.com/nekoimi/go-project-template/internal/storage/s3"
)

// New creates the configured storage backend. The factory lives in a
// subpackage because concrete drivers depend on the storage contracts package,
// and importing them from the root storage package would create an import
// cycle.
func New(cfg config.StorageConfig) (storage.FileStorage, error) {
	cfg.Normalize()

	switch strings.ToLower(cfg.Driver) {
	case "local":
		return local.New(cfg), nil
	case "s3":
		return s3.New(cfg.S3)
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", cfg.Driver)
	}
}
