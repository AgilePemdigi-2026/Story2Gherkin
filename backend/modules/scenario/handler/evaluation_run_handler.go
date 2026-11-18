package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
)

// EvaluationRunHandler exposes Run Evaluation endpoints.
type EvaluationRunHandler struct {
	service service.EvaluationRunService
}

func NewEvaluationRunHandler(s service.EvaluationRunService) *EvaluationRunHandler {
	return &EvaluationRunHandler{service: s}
}

// Create handles POST /api/v1/evaluation-runs.
func (h *EvaluationRunHandler) Create(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	var body dto.EvaluationRunCreateRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid evaluation request",
				"golden_dataset_id must be a valid UUID.",
				nil,
			),
		)
		return
	}

	created, err := h.service.Start(c.Request.Context(), userID, body.GoldenDatasetID)
	if err != nil {
		evaluationRunError(c, err)
		return
	}

	c.JSON(
		http.StatusCreated,
		utils.BuildResponseSuccess("Evaluation run created.", created),
	)
}

// Progress handles GET /api/v1/evaluation-runs/:run_id/progress.
func (h *EvaluationRunHandler) Progress(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	progress, err := h.service.Progress(c.Request.Context(), userID, c.Param("run_id"))
	if err != nil {
		evaluationRunError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess("Evaluation run progress.", progress),
	)
}

func evaluationRunError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "Evaluation request failed. Please try again."

	switch {
	case errors.Is(err, service.ErrEvaluationRunForbidden):
		status, message = http.StatusForbidden, "Only Quality Assurance users may run evaluations."
	case errors.Is(err, service.ErrEvaluationRunActive):
		status, message = http.StatusConflict, "Another evaluation run is still in progress."
	case errors.Is(err, service.ErrEvaluationRunNotFound):
		status, message = http.StatusNotFound, "Evaluation run or dataset not found."
	case errors.Is(err, service.ErrEvaluationRunEmptyDataset):
		status, message = http.StatusBadRequest, "The selected dataset has no items to evaluate."
	}

	c.AbortWithStatusJSON(
		status,
		utils.BuildResponseFailed(message, err.Error(), nil),
	)
}
