package handler

import (
	"net/http"
	"log"
	"errors"
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
	// Logout, sama Forgot Password
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
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, userDto.ErrEmailAlreadyExists) {
			res := utils.BuildResponseFailed(
				userDto.MESSAGE_FAILED_REGISTER_USER,
				"Email sudah terdaftar.",
				nil,
			)
			ctx.JSON(http.StatusConflict, res)
			return
		}
		log.Printf("register gagal: %v", err)

		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_REGISTER_USER,
			"Terjadi kesalahan pada server.",
			nil,
		)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(
		userDto.MESSAGE_SUCCESS_REGISTER_USER,
		result,
	)
	ctx.JSON(http.StatusCreated, res)
}

func (c *authHandler) Login(ctx *gin.Context) {
	// Respons berisi token tidak boleh disimpan oleh cache.
	ctx.Header("Cache-Control", "no-store")

	// Batasi ukuran request login menjadi 16 KiB.
	ctx.Request.Body = http.MaxBytesReader(
		ctx.Writer,
		ctx.Request.Body,
		16*1024,
	)

	var req authDto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_LOGIN,
			"Request harus berupa JSON dengan email valid dan password terisi.",
			nil,
		)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.authService.Login(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			res := utils.BuildResponseFailed(
				userDto.MESSAGE_FAILED_LOGIN,
				"Email atau password salah.",
				nil,
			)
			ctx.JSON(http.StatusUnauthorized, res)
			return
		}

		// Detail error internal tidak dikirim ke pengguna.
		res := utils.BuildResponseFailed(
			userDto.MESSAGE_FAILED_LOGIN,
			"Terjadi kesalahan pada server.",
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
