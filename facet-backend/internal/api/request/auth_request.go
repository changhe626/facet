package request

// LoginRequest defines the request body for user login
type LoginRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
	CaptchaID     string `json:"captchaId" binding:"required"`     // <-- added
	CaptchaAnswer string `json:"captchaAnswer" binding:"required"` // <-- added
}
