package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
	"regexp"
)

var evaluationReportCodePattern = regexp.MustCompile(`^ER-[0-9]{4,}$`)

type EvaluationReportHandler struct {
	service service.EvaluationReportService
}

func NewEvaluationReportHandler(
	s service.EvaluationReportService,
) *EvaluationReportHandler {
	return &EvaluationReportHandler{
		service: s,
	}
}

// evaluationReportError returns public messages without exposing internal errors.
func evaluationReportError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "Unable to process the evaluation report. Please try again."

	switch {
	case errors.Is(err, service.ErrEvaluationReportForbidden):
		status = http.StatusForbidden
		message = "Only Quality Assurance users can access evaluation reports."

	case errors.Is(err, service.ErrEvaluationReportInvalidRequest):
		status = http.StatusBadRequest
		message = "Invalid evaluation run ID or pagination parameters."

	case errors.Is(err, service.ErrEvaluationReportNotFound):
		status = http.StatusNotFound
		message = "Evaluation report not found."

	case errors.Is(err, service.ErrEvaluationReportNotReady):
		status = http.StatusConflict
		message = "CSV downloads are available only for completed evaluation runs."
	}

	c.AbortWithStatusJSON(
		status,
		utils.BuildResponseFailed(
			"Evaluation report request failed.",
			message,
			nil,
		),
	)
}

// List handles GET /api/v1/evaluation-runs.
func (h *EvaluationReportHandler) List(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		evaluationReportError(
			c,
			service.ErrEvaluationReportInvalidRequest,
		)
		return
	}

	pageSize, err := strconv.Atoi(
		c.DefaultQuery(
			"page_size",
			strconv.Itoa(service.EvaluationReportDefaultPageSize),
		),
	)
	if err != nil {
		evaluationReportError(
			c,
			service.ErrEvaluationReportInvalidRequest,
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
		evaluationReportError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Evaluation runs loaded.",
			result,
		),
	)
}

// Detail handles GET /api/v1/evaluation-runs/:run_id.
func (h *EvaluationReportHandler) Detail(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	report, err := h.service.Detail(
		c.Request.Context(),
		userID,
		c.Param("run_id"),
	)
	if err != nil {
		evaluationReportError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Evaluation report loaded.",
			report,
		),
	)
}

// Download handles GET /api/v1/evaluation-runs/:run_id/download.
// Download handles GET /api/v1/evaluation-runs/:run_id/download.
func (h *EvaluationReportHandler) Download(c *gin.Context) {
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	report, err := h.service.Download(
		c.Request.Context(),
		userID,
		c.Param("run_id"),
	)
	if err != nil {
		evaluationReportError(c, err)
		return
	}

	if !evaluationReportCodePattern.MatchString(report.ReportCode) {
		evaluationReportError(
			c,
			errors.New("invalid stored evaluation report code"),
		)
		return
	}

	data, err := service.BuildEvaluationReportCSV(report)
	if err != nil {
		evaluationReportError(c, err)
		return
	}

	filename := report.ReportCode + ".csv"

	c.Header(
		"Content-Disposition",
		`attachment; filename="`+filename+`"`,
	)

	c.Data(
		http.StatusOK,
		"text/csv; charset=utf-8",
		data,
	)
}
