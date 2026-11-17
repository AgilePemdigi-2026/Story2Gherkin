package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
	"github.com/story2gherkin/pkg/utils"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const maxGoldenUploadBytes int64 = 5 << 20

// GoldenDatasetHandler handles HTTP requests
// related to Golden Dataset management.
type GoldenDatasetHandler struct {
	service service.GoldenDatasetService
}

// Constructor
func NewGoldenDatasetHandler(
	s service.GoldenDatasetService,
) *GoldenDatasetHandler {
	return &GoldenDatasetHandler{
		service: s,
	}
}

// Upload handles POST /api/v1/golden-datasets.
func (h *GoldenDatasetHandler) Upload(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	// 1. Get authenticated user from JWT middleware
	userID := c.GetString("user_id")

	if userID == "" {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			utils.BuildResponseFailed(
				"unauthorized",
				"A valid Bearer token is required.",
				nil,
			),
		)
		return
	}

	// 2. Require multipart/form-data
	if c.ContentType() != "multipart/form-data" {
		c.AbortWithStatusJSON(
			http.StatusUnsupportedMediaType,
			utils.BuildResponseFailed(
				"invalid upload",
				"Content-Type must be multipart/form-data.",
				nil,
			),
		)
		return
	}

	// 3. Limit the entire HTTP request to 5 MiB
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxGoldenUploadBytes,
	)

	// 4. Read multipart form
	var form dto.GoldenDatasetUploadForm

	if err := c.ShouldBind(&form); err != nil {
		var sizeErr *http.MaxBytesError

		if errors.As(err, &sizeErr) {
			c.AbortWithStatusJSON(
				http.StatusRequestEntityTooLarge,
				utils.BuildResponseFailed(
					"invalid upload",
					"Upload exceeds the 5 MB request limit.",
					nil,
				),
			)
			return
		}

		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid upload",
				"Name and JSON file are required.",
				nil,
			),
		)
		return
	}

	// Remove temporary multipart files if any
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	raw, ok := readGoldenJSON(c, form.File)
	if !ok {
		return
	}

	// 8. Call Golden Dataset Service
	result, err := h.service.Upload(
		c.Request.Context(),
		userID,
		form.Name,
		form.Description,
		raw,
	)

	if err != nil {
		handleGoldenDatasetError(c, err)
		return
	}

	// 9. Return successful response
	c.JSON(
		http.StatusCreated,
		utils.BuildResponseSuccess(
			"Golden Dataset uploaded successfully.",
			result,
		),
	)
}

// Map service errors to HTTP responses.
func handleGoldenDatasetError(
	c *gin.Context,
	err error,
) {
	status := http.StatusInternalServerError
	message := "Golden Dataset operation failed."

	switch {
	case errors.Is(err, service.ErrGoldenDatasetForbidden):
		status = http.StatusForbidden
		message = "Only Quality Assurance users can access Golden Datasets."

	case errors.Is(err, service.ErrGoldenDatasetInvalid):
		status = http.StatusUnprocessableEntity
		message = err.Error()

	case errors.Is(err, service.ErrGoldenDatasetNotFound):
		status = http.StatusNotFound
		message = "Golden Dataset not found."

	case errors.Is(err, service.ErrGoldenDatasetLocked):
		status = http.StatusConflict
		message = "Golden Dataset has already been used in an evaluation and cannot be changed."

	case errors.Is(err, service.ErrUnavailable):
		status = http.StatusServiceUnavailable
		message = "Gherkin validator is unavailable."

	case errors.Is(err, service.ErrTimeout):
		status = http.StatusGatewayTimeout
		message = "Gherkin validation timed out."

	case errors.Is(err, service.ErrBadGateway):
		status = http.StatusBadGateway
		message = "Gherkin validator returned an invalid response."
	}

	c.AbortWithStatusJSON(
		status,
		utils.BuildResponseFailed(
			"Golden Dataset operation failed",
			message,
			nil,
		),
	)
}

func readGoldenJSON(
	c *gin.Context,
	header *multipart.FileHeader,
) ([]byte, bool) {

	if header == nil ||
		!strings.EqualFold(
			filepath.Ext(header.Filename),
			".json",
		) {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid upload",
				"Only .json files are supported.",
				nil,
			),
		)
		return nil, false
	}

	file, err := header.Open()
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid upload",
				"Unable to read uploaded file.",
				nil,
			),
		)
		return nil, false
	}
	defer file.Close()

	raw, err := io.ReadAll(
		io.LimitReader(file, maxGoldenUploadBytes+1),
	)

	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid upload",
				"Unable to read JSON content.",
				nil,
			),
		)
		return nil, false
	}

	if int64(len(raw)) > maxGoldenUploadBytes {
		c.AbortWithStatusJSON(
			http.StatusRequestEntityTooLarge,
			utils.BuildResponseFailed(
				"invalid upload",
				"JSON file exceeds the size limit.",
				nil,
			),
		)
		return nil, false
	}

	return raw, true
}

// Update handles PUT /api/v1/golden-datasets/:id.
func (h *GoldenDatasetHandler) Update(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	if c.ContentType() != "multipart/form-data" {
		c.AbortWithStatusJSON(
			http.StatusUnsupportedMediaType,
			utils.BuildResponseFailed(
				"invalid update",
				"Content-Type must be multipart/form-data.",
				nil,
			),
		)
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxGoldenUploadBytes,
	)

	var form dto.GoldenDatasetUpdateForm

	if err := c.ShouldBind(&form); err != nil {
		var sizeErr *http.MaxBytesError

		if errors.As(err, &sizeErr) {
			c.AbortWithStatusJSON(
				http.StatusRequestEntityTooLarge,
				utils.BuildResponseFailed(
					"invalid update",
					"Upload exceeds the 5 MB request limit.",
					nil,
				),
			)
			return
		}

		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid update",
				"Name is required.",
				nil,
			),
		)
		return
	}

	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}

	var raw []byte

	if form.File != nil {
		raw, ok = readGoldenJSON(c, form.File)
		if !ok {
			return
		}
	}

	result, err := h.service.Update(
		c.Request.Context(),
		userID,
		c.Param("id"),
		form.Name,
		form.Description,
		raw,
	)

	if err != nil {
		handleGoldenDatasetError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Golden Dataset updated successfully.",
			result,
		),
	)
}

// Delete handles DELETE /api/v1/golden-datasets/:id.
func (h *GoldenDatasetHandler) Delete(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		userID,
		c.Param("id"),
	); err != nil {
		handleGoldenDatasetError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Golden Dataset deleted successfully.",
			nil,
		),
	)
}

// List handles GET /api/v1/golden-datasets.
func (h *GoldenDatasetHandler) List(c *gin.Context) {
	// 1. Get authenticated user.
	// Reuses the helper from review_handler.go.
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	// 2. Default pagination.
	page := 1
	pageSize := 20

	// 3. Parse page parameter.
	if raw := c.Query("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)

		if err != nil || parsed < 1 || parsed > 1000000 {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				utils.BuildResponseFailed(
					"invalid pagination",
					"Page must be a positive integer.",
					nil,
				),
			)
			return
		}

		page = parsed
	}

	// 4. Parse page_size parameter.
	if raw := c.Query("page_size"); raw != "" {
		parsed, err := strconv.Atoi(raw)

		if err != nil || parsed < 1 || parsed > 100 {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				utils.BuildResponseFailed(
					"invalid pagination",
					"Page size must be between 1 and 100.",
					nil,
				),
			)
			return
		}

		pageSize = parsed
	}

	// 5. Read optional search keyword.
	search := strings.TrimSpace(c.Query("search"))

	if len(search) > 200 {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid search",
				"Search keyword is too long.",
				nil,
			),
		)
		return
	}

	// 6. Call service.
	result, err := h.service.List(
		c.Request.Context(),
		userID,
		search,
		page,
		pageSize,
	)

	if err != nil {
		handleGoldenDatasetError(c, err)
		return
	}

	// 7. Return list response.
	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Golden Datasets retrieved successfully.",
			result,
		),
	)
}

// Detail handles GET /api/v1/golden-datasets/:id.
func (h *GoldenDatasetHandler) Detail(c *gin.Context) {
	// 1. Get authenticated user.
	userID, ok := reviewUser(c)
	if !ok {
		return
	}

	// 2. Read dataset ID from URL.
	datasetID := c.Param("id")

	if datasetID == "" {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			utils.BuildResponseFailed(
				"invalid dataset ID",
				"Golden Dataset ID is required.",
				nil,
			),
		)
		return
	}

	// 3. Call service.
	result, err := h.service.GetDetail(
		c.Request.Context(),
		userID,
		datasetID,
	)

	if err != nil {
		handleGoldenDatasetError(c, err)
		return
	}

	// 4. Return dataset and its items.
	c.JSON(
		http.StatusOK,
		utils.BuildResponseSuccess(
			"Golden Dataset detail retrieved successfully.",
			result,
		),
	)
}
