package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/service"
)

type UploadHandler struct {
	fileService service.FileService
	logger      *zap.Logger
}

func NewUploadHandler(fileService service.FileService, logger *zap.Logger) *UploadHandler {
	return &UploadHandler{fileService: fileService, logger: logger}
}

// UploadSingle godoc
// @Summary      Upload a single file
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file   formData  file    true  "File to upload"
// @Param        folder formData  string  false "Upload folder"
// @Success      200    {object}  resp.JsonResponse
// @Failure      400    {object}  resp.JsonResponse
// @Router       /upload/single [post]
func (h *UploadHandler) UploadSingle(c *gin.Context) (any, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "missing file")
	}

	folder := c.DefaultPostForm("folder", "uploads")

	return h.fileService.UploadSingle(c.Request.Context(), file, folder)
}

// UploadMultiple godoc
// @Summary      Upload multiple files
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        files  formData  []file  true  "Files to upload"
// @Param        folder formData  string  false "Upload folder"
// @Success      200    {object}  resp.JsonResponse
// @Failure      400    {object}  resp.JsonResponse
// @Router       /upload/multiple [post]
func (h *UploadHandler) UploadMultiple(c *gin.Context) (any, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "invalid multipart form")
	}

	files := form.File["files"]
	if len(files) == 0 {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "no files provided")
	}

	folder := c.DefaultPostForm("folder", "uploads")

	return h.fileService.UploadMultiple(c.Request.Context(), files, folder)
}
