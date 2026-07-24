package validator

import (
	"errors"
	"facet-backend/internal/api/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// translations maps validator tags to English.
var translations = map[string]string{
	"required": "is required",
	"email":    "must be a valid email address",
	"min":      "must be at least %s characters long",
	"max":      "must be at most %s characters long",
}

// HandleBindingError checks for validation errors and sends a translated response.
// It returns true if an error was handled, false otherwise.
func HandleBindingError(c *gin.Context, err error) bool {
	var verr validator.ValidationErrors
	if !errors.As(err, &verr) {
		// This is not a validation error, likely a malformed JSON.
		response.CustomError(c, http.StatusBadRequest, "invalid request format")
		return true
	}

	// Just translate the first error for a concise message.
	fieldErr := verr[0]
	fieldName := fieldErr.Field()
	tag := fieldErr.Tag()

	translation, ok := translations[tag]
	if !ok {
		// Fallback for untranslated tags
		response.CustomError(c, http.StatusBadRequest, fmt.Sprintf("field %s validation failed, rule: %s", fieldName, tag))
		return true
	}

	message := fmt.Sprintf("field %s %s", fieldName, translation)
	if fieldErr.Param() != "" {
		message = fmt.Sprintf("field %s %s", fieldName, fmt.Sprintf(translation, fieldErr.Param()))
	}

	response.CustomError(c, http.StatusBadRequest, message)
	return true
}
