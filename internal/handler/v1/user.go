package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/response"
	"github.com/nekoimi/go-project-template/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized)
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		response.ErrorWithMsg(c, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}

	response.Success(c, profile)
}
