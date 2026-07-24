package handler

import (
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/api/validator"
	"facet-backend/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService    service.AuthService
	captchaService service.CaptchaService
}

func NewAuthHandler(authService service.AuthService, captchaService service.CaptchaService) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		captchaService: captchaService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("login request bind failed", "error", err)
		validator.HandleBindingError(c, err)
		return
	}

	if !h.captchaService.Verify(req.CaptchaID, req.CaptchaAnswer) {
		response.CustomError(c, http.StatusUnauthorized, "invalid captcha")
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			slog.Warn("authentication failed", "email", req.Email)
			response.Error(c, http.StatusUnauthorized, service.ErrInvalidCredentials)
			return
		}

		slog.Error("internal error during login", "email", req.Email, "error", err)
		response.CustomError(c, http.StatusInternalServerError, "an internal error occurred")
		return
	}

	response.Success(c, resp)
}
