package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
)

type Handler struct {
	userService Service
	logger      *zap.Logger
}

func NewHandler(userService Service, logger *zap.Logger) *Handler {
	return &Handler{userService: userService, logger: logger}
}

// GetProfile godoc
// @Summary      Get current user profile
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resp.JsonResponse{data=user.UserResponse}
// @Failure      401  {object}  resp.JsonResponse
// @Failure      404  {object}  resp.JsonResponse
// @Router       /users/profile [get]
func (h *Handler) GetProfile(c *gin.Context) (any, error) {
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
