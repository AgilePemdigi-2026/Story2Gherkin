package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/pkg/utils"
)

func Authenticate(jwtService service.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := strings.Fields(ctx.GetHeader("Authorization"))
		if len(authHeader) == 0 {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_FOUND, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if len(authHeader) != 2 || !strings.EqualFold(authHeader[0], "Bearer") {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		rawToken := authHeader[1]
		userID, err := jwtService.GetUserIDByToken(rawToken)
		if err != nil || userID == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		ctx.Set("user_id", userID)
		ctx.Next()
	}
}
