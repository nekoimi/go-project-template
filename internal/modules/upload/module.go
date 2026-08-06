package upload

import (
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/pkg/resp"
)

func init() {
	framework.Register(NewModule(), framework.ScopeHTTP)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "upload"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) {
		return nil
	}

	fileService := NewService(
		ctx.Storage,
		ctx.Config.Storage.Upload.MaxFileSize,
		ctx.Config.Storage.Upload.AllowedExts,
		ctx.Config.Storage.Upload.AllowedMIMEs,
	)
	uploadHandler := NewHandler(fileService, ctx.Logger)

	upload := ctx.Protected.Group("/upload")
	upload.POST("/single", resp.Handle(uploadHandler.UploadSingle, ctx.Logger))
	upload.POST("/multiple", resp.Handle(uploadHandler.UploadMultiple, ctx.Logger))

	return nil
}
