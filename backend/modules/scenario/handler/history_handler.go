package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
)

type HistoryHandler struct {
	service service.HistoryService
}

func NewHistoryHandler(
	s service.HistoryService,
) *HistoryHandler {
	return &HistoryHandler{service: s}
}

// Menangani error History sesuai HTTP status.
func historyError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "History request failed. Please try again."

	switch {
	case errors.Is(err, service.ErrHistoryForbidden):
		status = http.StatusForbidden
		message = "Only System Analysts may access generation history."

	case errors.Is(err, service.ErrHistoryNotFound):
		status = http.StatusNotFound
		message = "History not found."

	case errors.Is(err, service.ErrHistoryInvalidPagination):
		status = http.StatusBadRequest
		message = "Invalid pagination parameters."
	}

	c.AbortWithStatusJSON(
		status,
		utils.BuildResponseFailed(
			"failed history request",
			message,
			nil,
		),
	)
}

// GET /api/v1/history
func (h *HistoryHandler) List(c *gin.Context) {
	// Memakai helper autentikasi dari review_handler.go.
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	page, err := strconv.Atoi(
		c.DefaultQuery("page", "1"),
	)

	if err != nil || page < 1 || page > 1000000 {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid pagination",
				"page must be between 1 and 1000000",
				nil,
			),
		)
		return
	}

	pageSize, err := strconv.Atoi(
		c.DefaultQuery("page_size", "20"),
	)

	if err != nil || pageSize < 1 || pageSize > 100 {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid pagination",
				"page_size must be between 1 and 100",
				nil,
			),
		)
		return
	}

	result, err := h.service.List(
		c.Request.Context(),
		userID,
		page,
		pageSize,
	)

	if err != nil {
		historyError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Generation history retrieved successfully.",
			result,
		),
	)
}

// GET /api/v1/history/:id
func (h *HistoryHandler) Detail(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	result, err := h.service.Detail(
		c.Request.Context(),
		userID,
		c.Param("id"),
	)

	if err != nil {
		historyError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Generation history detail retrieved successfully.",
			result,
		),
	)
}
