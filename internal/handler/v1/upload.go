package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/response"
	"github.com/nekoimi/go-project-template/internal/service"
)

type UploadHandler struct {
	fileService service.FileService
}

func NewUploadHandler(fileService service.FileService) *UploadHandler {
	return &UploadHandler{fileService: fileService}
}

func (h *UploadHandler) UploadSingle(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithMsg(c, http.StatusBadRequest, errcode.BadRequest, "missing file")
		return
	}

	folder := c.DefaultPostForm("folder", "uploads")

	result, err := h.fileService.UploadSingle(c.Request.Context(), file, folder)
	if err != nil {
		response.ErrorWithMsg(c, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.ErrorWithMsg(c, http.StatusBadRequest, errcode.BadRequest, "invalid multipart form")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.ErrorWithMsg(c, http.StatusBadRequest, errcode.BadRequest, "no files provided")
		return
	}

	folder := c.DefaultPostForm("folder", "uploads")

	results, err := h.fileService.UploadMultiple(c.Request.Context(), files, folder)
	if err != nil {
		response.ErrorWithMsg(c, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}

	response.Success(c, results)
}
