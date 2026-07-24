package middleware

import (
	"errors"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler creates a global error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			switch {
			// --- Evaluation API specific errors (400 Bad Request) ---
			case errors.Is(err, domain.ErrMissingSDKKey),
				errors.Is(err, domain.ErrMissingFeatureName),
				errors.Is(err, domain.ErrMissingAccountID),
				errors.Is(err, domain.ErrMissingAppName):
				response.Error(c, http.StatusBadRequest, err)

			// --- Forbidden errors ---
			case errors.Is(err, domain.ErrTenantInactive),
				errors.Is(err, domain.ErrEnvironmentInactive),
				errors.Is(err, domain.ErrFeatureFlagDisabled),
				errors.Is(err, domain.ErrFeatureFlagArchived):
				response.Error(c, http.StatusForbidden, err)

			// --- Evaluation API specific errors (404 Not Found) ---
			case errors.Is(err, domain.ErrEnvironmentInvalidSDKKey):
				response.Error(c, http.StatusNotFound, err)

			// --- Generic resource not found errors ---
			case errors.Is(err, domain.ErrEnvironmentNotFound),
				errors.Is(err, domain.ErrUserNotFound),
				errors.Is(err, domain.ErrTenantNotFound),
				errors.Is(err, domain.ErrFeatureFlagNotFound),
				errors.Is(err, domain.ErrSegmentNotFound),
				errors.Is(err, domain.ErrTenantSuperAdminNotFound),
				errors.Is(err, domain.ErrTenantLevelNotFound):
				response.Error(c, http.StatusNotFound, err)

			// --- Generic uniqueness constraint conflict errors ---
			case errors.Is(err, domain.ErrEnvironmentNameExists),
				errors.Is(err, domain.ErrEmailExistsInTenant),
				errors.Is(err, domain.ErrFeatureFlagNameExistsInTenant),
				errors.Is(err, domain.ErrSegmentNameExists),
				errors.Is(err, domain.ErrTenantNameExists):
				response.Error(c, http.StatusConflict, err)

			// --- Validation / bad request errors ---
			case errors.Is(err, domain.ErrInvalidContextFormat),
				errors.Is(err, domain.ErrUnknownSegmentOperator),
				errors.Is(err, domain.ErrInvalidInput),
				errors.Is(err, domain.ErrRoleInTenant):
				response.Error(c, http.StatusBadRequest, err)

			// --- Default handling ---
			default:
				slog.Error("internal server error", "error", err.Error(), "path", c.Request.URL.Path)
				response.CustomError(c, http.StatusInternalServerError, "an internal error occurred")
			}

			c.Abort()
		}
	}
}
