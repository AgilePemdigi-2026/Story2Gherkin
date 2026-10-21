package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/pkg/constants"
	"github.com/story2gherkin/pkg/utils"
	"gorm.io/gorm"
)

type authHandler struct {
	authService service.AuthService
	db             *gorm.DB
}

type AuthHandler interface {
	Register(ctx *gin.Context)
	// Login, Logout, sama Forgot Password
}

func NewAuthController(injector *do.Injector, as service.AuthService) AuthHandler {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &authHandler{
		authService:    as,
		db:             db,
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
		res := utils.BuildResponseFailed(userDto.MESSAGE_FAILED_REGISTER_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(userDto.MESSAGE_SUCCESS_REGISTER_USER, result)
	ctx.JSON(http.StatusOK, res)
}