package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/service"
)

type fakeRunService struct {
	created  dto.EvaluationRunCreated
	startErr error
}

func (f fakeRunService) Start(context.Context, string, string) (dto.EvaluationRunCreated, error) {
	return f.created, f.startErr
}

func (f fakeRunService) Progress(context.Context, string, string) (dto.EvaluationRunProgress, error) {
	return dto.EvaluationRunProgress{}, nil
}

func createContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set("user_id", uuid.NewString())
	c.Request = httptest.NewRequest(http.MethodPost, "/evaluation-runs", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, recorder
}

func TestCreateMapsActiveRunToConflict(t *testing.T) {
	c, recorder := createContext(`{"golden_dataset_id":"` + uuid.NewString() + `"}`)
	h := handler.NewEvaluationRunHandler(fakeRunService{startErr: service.ErrEvaluationRunActive})

	h.Create(c)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateRejectsInvalidBody(t *testing.T) {
	c, recorder := createContext(`{"golden_dataset_id":"not-a-uuid"}`)
	h := handler.NewEvaluationRunHandler(fakeRunService{})

	h.Create(c)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateReturnsCreatedRun(t *testing.T) {
	c, recorder := createContext(`{"golden_dataset_id":"` + uuid.NewString() + `"}`)
	h := handler.NewEvaluationRunHandler(fakeRunService{created: dto.EvaluationRunCreated{
		RunID:      "abc",
		ReportCode: "ER-0001",
		Status:     entities.EvaluationRunQueued,
		TotalCases: 3,
	}})

	h.Create(c)

	if recorder.Code != http.StatusCreated || !strings.Contains(recorder.Body.String(), `"run_id":"abc"`) {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
}
