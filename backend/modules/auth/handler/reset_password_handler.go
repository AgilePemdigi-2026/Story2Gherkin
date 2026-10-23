package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	authDto "github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/pkg/utils"
)

func (c *authHandler) ResetPassword(ctx *gin.Context) {
	var req authDto.ResetPasswordRequest

	// 1. Validasi request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(
			"Invalid reset password request",
			"Token and a valid new password are required",
			nil,
		)

		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// 2. Proses perubahan password
	err := c.authService.ResetPassword(
		ctx.Request.Context(),
		req.Token,
		req.NewPassword,
	)

	// 3. Tangani token tidak valid
	if errors.Is(err, service.ErrInvalidResetToken) {
		res := utils.BuildResponseFailed(
			"Password reset failed",
			"Reset link is invalid, expired, or already used",
			nil,
		)

		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// 4. Tangani password tidak valid
	if errors.Is(err, service.ErrInvalidNewPassword) {
		res := utils.BuildResponseFailed(
			"Invalid new password",
			err.Error(),
			nil,
		)

		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// 5. Tangani error internal
	if err != nil {
		log.Printf("Reset password internal error: %v", err)

		res := utils.BuildResponseFailed(
			"Password reset failed",
			"Unable to reset password. Please try again",
			nil,
		)

		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	// 6. Respons berhasil
	res := utils.BuildResponseSuccess(
		"Password has been reset successfully",
		nil,
	)

	ctx.JSON(http.StatusOK, res)
}
