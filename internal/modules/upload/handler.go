package upload

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
)

type Handler struct {
	fileService     Service
	logger          *zap.Logger
	maxRequestBytes int64
	maxFiles        int
}

func NewHandler(fileService Service, logger *zap.Logger, maxRequestSizeMB, maxFiles int) *Handler {
	return &Handler{
		fileService:     fileService,
		logger:          logger,
		maxRequestBytes: int64(maxRequestSizeMB) * 1024 * 1024,
		maxFiles:        maxFiles,
	}
}

func (h *Handler) limitRequestBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxRequestBytes)
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
func (h *Handler) UploadSingle(c *gin.Context) (any, error) {
	h.limitRequestBody(c)
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
func (h *Handler) UploadMultiple(c *gin.Context) (any, error) {
	h.limitRequestBody(c)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "invalid multipart form")
	}

	files := form.File["files"]
	if len(files) == 0 {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "no files provided")
	}
	if len(files) > h.maxFiles {
		return nil, errcode.NewWithDetail(errcode.BadRequest, "too many files")
	}

	folder := c.DefaultPostForm("folder", "uploads")

	return h.fileService.UploadMultiple(c.Request.Context(), files, folder)
}
