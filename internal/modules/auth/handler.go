package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
)

type Handler struct {
	authService Service
	logger      *zap.Logger
}

func NewHandler(authService Service, logger *zap.Logger) *Handler {
	return &Handler{authService: authService, logger: logger}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      auth.RegisterRequest  true  "Register request"
// @Success      200   {object}  resp.JsonResponse{data=auth.AuthResponse}
// @Failure      400   {object}  resp.JsonResponse
// @Failure      409   {object}  resp.JsonResponse
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) (any, error) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.NewWithDetail(errcode.Validation, err.Error())
	}

	return h.authService.Register(c.Request.Context(), req)
}

// Login godoc
// @Summary      User login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      auth.LoginRequest  true  "Login request"
// @Success      200   {object}  resp.JsonResponse{data=auth.AuthResponse}
// @Failure      400   {object}  resp.JsonResponse
// @Failure      401   {object}  resp.JsonResponse
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) (any, error) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.NewWithDetail(errcode.Validation, err.Error())
	}

	return h.authService.Login(c.Request.Context(), req)
}
