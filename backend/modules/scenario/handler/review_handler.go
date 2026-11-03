package handler

import (
	"errors"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
)

type ReviewHandler struct{ service service.ReviewService }

func NewReviewHandler(s service.ReviewService) *ReviewHandler { return &ReviewHandler{service: s} }

func reviewUser(c *gin.Context) (string, bool) {
	c.Header("Cache-Control", "no-store")
	userID, _ := c.Get("user_id")
	id, ok := userID.(string)
	if !ok || id == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.BuildResponseFailed("unauthorized", "A valid Bearer token is required.", nil))
		return "", false
	}
	return id, true
}

func reviewError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "Review request failed. Please try again."
	var data any
	var invalid *service.InvalidCorrectionError
	switch {
	case errors.As(err, &invalid):
		status, message = http.StatusUnprocessableEntity, invalid.Error()
		data = gin.H{"validation": invalid.Validation}
	case errors.Is(err, service.ErrReviewForbidden):
		status, message = http.StatusForbidden, "Only System Analysts may review, correct, or export scenarios."
	case errors.Is(err, service.ErrScenarioNotFound):
		status, message = http.StatusNotFound, "Scenario not found."
	case errors.Is(err, repository.ErrCorrectionConflict):
		status, message = http.StatusConflict, "This scenario changed in another session. Reload it before saving again."
	case errors.Is(err, service.ErrInvalidRequest):
		status, message = http.StatusBadRequest, "Provide Gherkin text and the scenario's updated_at value."
	case errors.Is(err, service.ErrUnavailable):
		status, message = http.StatusServiceUnavailable, "Gherkin validator is unavailable. Your saved scenario has not changed."
	case errors.Is(err, service.ErrTimeout):
		status, message = http.StatusGatewayTimeout, "Gherkin validation timed out. Please retry."
	case errors.Is(err, service.ErrBadGateway):
		status, message = http.StatusBadGateway, "Gherkin validator returned an invalid response."
	}
	c.AbortWithStatusJSON(status, utils.BuildResponseFailed("failed review scenario", message, data))
}

func (h *ReviewHandler) List(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		reviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success list scenarios", items))
}

func (h *ReviewHandler) Detail(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}
	item, err := h.service.Detail(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		reviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success get scenario", item))
}

func (h *ReviewHandler) Save(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 512*1024)
	var input dto.SaveCorrectionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		status := http.StatusBadRequest
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			status = http.StatusRequestEntityTooLarge
		}
		c.AbortWithStatusJSON(status, utils.BuildResponseFailed("invalid correction", "Provide Gherkin text (maximum 65536 characters) and updated_at.", nil))
		return
	}
	item, err := h.service.Save(c.Request.Context(), userID, c.Param("id"), input)
	if err != nil {
		reviewError(c, err)
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("Correction saved successfully!", item))
}

func (h *ReviewHandler) Download(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}
	item, err := h.service.Download(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		reviewError(c, err)
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": service.FeatureFilename(item.FeatureName)}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(item.Gherkin))
}
