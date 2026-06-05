package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/model"
	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterRequest  true  "Register request"
// @Success      200   {object}  response.APIResponse{data=model.AuthResponse}
// @Failure      400   {object}  response.APIResponse
// @Failure      409   {object}  response.APIResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) (any, error) {
	var req model.RegisterRequest
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
// @Param        body  body      model.LoginRequest  true  "Login request"
// @Success      200   {object}  response.APIResponse{data=model.AuthResponse}
// @Failure      400   {object}  response.APIResponse
// @Failure      401   {object}  response.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) (any, error) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errcode.NewWithDetail(errcode.Validation, err.Error())
	}

	return h.authService.Login(c.Request.Context(), req)
}
