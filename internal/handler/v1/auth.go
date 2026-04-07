package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nekoimi/go-project-template/internal/dto"
	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
	"github.com/nekoimi/go-project-template/internal/pkg/response"
	"github.com/nekoimi/go-project-template/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, http.StatusBadRequest, errcode.BadRequest, err.Error())
		return
	}

	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.ErrorWithMsg(c, http.StatusConflict, errcode.Conflict, "email already exists")
			return
		}
		if errors.Is(err, service.ErrUsernameAlreadyExists) {
			response.ErrorWithMsg(c, http.StatusConflict, errcode.Conflict, "username already exists")
			return
		}
		response.ErrorWithMsg(c, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, http.StatusBadRequest, errcode.BadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.ErrorWithMsg(c, http.StatusUnauthorized, errcode.Unauthorized, "invalid email or password")
			return
		}
		response.ErrorWithMsg(c, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}

	response.Success(c, result)
}
