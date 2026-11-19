package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	authDto "github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/pkg/constants"
	"github.com/story2gherkin/pkg/utils"
	"gorm.io/gorm"
)

type authHandler struct {
	authService service.AuthService
	db          *gorm.DB
}

type AuthHandler interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	Logout(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
	SendPasswordReset(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)
}

func NewAuthController(
	injector *do.Injector,
	as service.AuthService,
) AuthHandler {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)

	return &authHandler{
		authService: as,
		db:          db,
	}
}

func (c *authHandler) Register(ctx *gin.Context) {
	var req userDto.UserCreateRequest

	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY,
			err.Error(),
			nil,
		)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Register(
		ctx.Request.Context(),
		req,
	)

	if err != nil {
		if errors.Is(err, userDto.ErrEmailAlreadyExists) {
			res := utils.BuildResponseFailed(
				userDto.MESSAGE_FAILED_REGISTER_USER,
				"This email address is already registered.",
				nil,
			)
			ctx.JSON(http.StatusConflict, res)
			return
		}

		log.Printf("register gagal: %v", err)

		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_REGISTER_USER,
			"An internal server error occurred.",
			nil,
		)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	if errors.Is(err, authDto.ErrInvalidNewPassword) {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				userDto.MESSAGE_FAILED_REGISTER_USER,
				authDto.ErrInvalidNewPassword.Error(),
				nil,
			),
		)
		return
	}

	res := utils.BuildResponseSuccess(
		userDto.MESSAGE_SUCCESS_REGISTER_USER,
		result,
	)

	ctx.JSON(http.StatusCreated, res)
}

func (c *authHandler) Login(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")

	ctx.Request.Body = http.MaxBytesReader(
		ctx.Writer,
		ctx.Request.Body,
		16*1024,
	)

	var req userDto.UserLoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_LOGIN,
			"The request must contain JSON with a valid email address and a non-empty password.",
			nil,
		)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Login(
		ctx.Request.Context(),
		req,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			res := utils.BuildResponseFailed(
				userDto.MESSAGE_FAILED_LOGIN,
				"Incorrect email or password.",
				nil,
			)
			ctx.JSON(http.StatusUnauthorized, res)
			return
		}

		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_LOGIN,
			"An internal server error occurred.",
			nil,
		)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(
		userDto.MESSAGE_SUCCESS_LOGIN,
		result,
	)

	ctx.JSON(http.StatusOK, res)
}

func (c *authHandler) Logout(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")

	res := utils.BuildResponseSuccess(
		"success logout",
		nil,
	)

	ctx.JSON(http.StatusOK, res)
}

func (c *authHandler) RefreshToken(ctx *gin.Context) {
	var req authDto.RefreshTokenRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.RefreshToken(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(authDto.MESSAGE_FAILED_REFRESH_TOKEN, err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, res)
		return
	}

	res := utils.BuildResponseSuccess(authDto.MESSAGE_SUCCESS_REFRESH_TOKEN, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *authHandler) SendPasswordReset(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")

	ctx.Request.Body = http.MaxBytesReader(
		ctx.Writer,
		ctx.Request.Body,
		16*1024,
	)

	var req authDto.SendPasswordResetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				authDto.MESSAGE_FAILED_SEND_PASSWORD_RESET,
				"Please provide a valid email address.",
				nil,
			),
		)
		return
	}

	if err := c.authService.SendPasswordReset(
		ctx.Request.Context(),
		req,
	); err != nil {
		log.Printf("password reset email failed: %v", err)

		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			utils.BuildResponseFailed(
				authDto.MESSAGE_FAILED_SEND_PASSWORD_RESET,
				"Unable to process your password reset request. Please try again later.",
				nil,
			),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"If an account exists for this email, a password reset link will be sent.",
			nil,
		),
	)
}

func (c *authHandler) ResetPassword(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")

	ctx.Request.Body = http.MaxBytesReader(
		ctx.Writer,
		ctx.Request.Body,
		16*1024,
	)

	var req authDto.ResetPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				authDto.MESSAGE_FAILED_RESET_PASSWORD,
				"Please provide a reset token and a valid new password.",
				nil,
			),
		)
		return
	}

	err := c.authService.ResetPassword(
		ctx.Request.Context(),
		req,
	)

	if err != nil {
		status := http.StatusInternalServerError
		message := "Unable to reset your password. Please try again later."

		switch {
		case errors.Is(err, authDto.ErrPasswordResetToken):
			status = http.StatusBadRequest
			message = "This password reset link is invalid or has expired. Please request a new link."

		case errors.Is(err, authDto.ErrInvalidNewPassword):
			status = http.StatusBadRequest
			message = authDto.ErrInvalidNewPassword.Error()

		default:
			log.Printf("password reset failed: %v", err)
		}

		ctx.AbortWithStatusJSON(
			status,
			utils.BuildResponseFailed(
				authDto.MESSAGE_FAILED_RESET_PASSWORD,
				message,
				nil,
			),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Your password has been reset. Please sign in with your new password.",
			nil,
		),
	)
}
