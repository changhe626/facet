package handler

import (
	"facet-backend/internal/api/response"
	"facet-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct {
	captchaService service.CaptchaService
}

func NewCaptchaHandler(captchaService service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{captchaService: captchaService}
}

func (h *CaptchaHandler) GetCaptcha(c *gin.Context) {
	captcha, err := h.captchaService.Generate()
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, captcha)
}
