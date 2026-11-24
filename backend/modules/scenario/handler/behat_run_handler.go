package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
)

// BehatRunHandler exposes end-to-end test endpoints for saved scenarios.
type BehatRunHandler struct {
	service service.BehatRunService
}

func NewBehatRunHandler(s service.BehatRunService) *BehatRunHandler {
	return &BehatRunHandler{service: s}
}

// Run handles POST /api/v1/scenarios/:id/behat-runs.
func (h *BehatRunHandler) Run(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	var body dto.BehatRunRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid end-to-end test request",
				"Provide a base URL and at least one mapped step.",
				nil,
			),
		)
		return
	}

	result, err := h.service.Run(c.Request.Context(), userID, c.Param("id"), body)
	if err != nil {
		behatRunError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponseSuccess("End-to-end test finished.", result))
}

// History handles GET /api/v1/scenarios/:id/behat-runs.
func (h *BehatRunHandler) History(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	runs, err := h.service.History(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		behatRunError(c, err)
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponseSuccess("End-to-end test history.", runs))
}

func behatRunError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "End-to-end test failed. Please try again."

	switch {
	case errors.Is(err, service.ErrBehatForbidden):
		status, message = http.StatusForbidden, "Only System Analysts may run end-to-end tests."
	case errors.Is(err, service.ErrBehatNotFound):
		status, message = http.StatusNotFound, "Scenario not found."
	case errors.Is(err, service.ErrBehatInvalid):
		status, message = http.StatusBadRequest, "Provide a valid base URL and mapped steps."
	}

	c.AbortWithStatusJSON(
		status,
		utils.BuildResponseFailed(message, err.Error(), nil),
	)
}
