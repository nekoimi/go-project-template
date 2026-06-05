package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/service"
)

type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

// GetProfile godoc
// @Summary      Get current user profile
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=model.UserResponse}
// @Failure      401  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Router       /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) (any, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return nil, errcode.New(errcode.Unauthorized)
	}

	uidStr, ok := userID.(string)
	if !ok || uidStr == "" {
		return nil, errcode.New(errcode.Unauthorized)
	}

	return h.userService.GetProfile(c.Request.Context(), uidStr)
}
