package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
)

type ScenarioHandler interface{ Generate(*gin.Context) }
type scenarioHandler struct{ service service.ScenarioService }

func NewScenarioHandler(s service.ScenarioService) ScenarioHandler {
	return &scenarioHandler{service: s}
}

func (h *scenarioHandler) Generate(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status := http.StatusBadRequest
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			status = http.StatusRequestEntityTooLarge
		}
		c.AbortWithStatusJSON(status, utils.BuildResponseFailed("failed generate scenarios", "Invalid JSON request or request exceeds the size limit.", nil))
		return
	}
	result, err := h.service.Generate(c.Request.Context(), req)
	if err != nil {
		status, message := http.StatusInternalServerError, "Scenario generation failed."
		switch {
		case errors.Is(err, service.ErrInvalidRequest):
			status, message = http.StatusBadRequest, "Provide a valid user story, acceptance criteria, language, and scenario limit."
		case errors.Is(err, service.ErrUnavailable):
			status, message = http.StatusServiceUnavailable, "Scenario generator is unavailable or busy."
		case errors.Is(err, service.ErrTimeout):
			status, message = http.StatusGatewayTimeout, "Scenario generation timed out."
		case errors.Is(err, service.ErrBadGateway):
			status, message = http.StatusBadGateway, "Scenario generator returned an invalid result."
		}
		c.JSON(status, utils.BuildResponseFailed("failed generate scenarios", message, nil))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success generate scenarios", result))
}
