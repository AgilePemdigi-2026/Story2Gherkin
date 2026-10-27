package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

type stubGenerator struct {
	calls int
	fn    func(ctx context.Context, req dto.GenerateRequest) (dto.GenerateResponse, error)
}

func (s *stubGenerator) Generate(ctx context.Context, req dto.GenerateRequest) (dto.GenerateResponse, error) {
	s.calls++
	return s.fn(ctx, req)
}

type memReqRepo struct {
	store        map[uuid.UUID]entities.GenerationRequest
	creates      int
	updates      int
	createErr    error
}

func (m *memReqRepo) Create(_ context.Context, _ *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error) {
	if m.createErr != nil {
		return entities.GenerationRequest{}, m.createErr
	}
	if m.store == nil {
		m.store = map[uuid.UUID]entities.GenerationRequest{}
	}
	m.store[req.ID] = req
	m.creates++
	return req, nil
}

func (m *memReqRepo) Update(_ context.Context, _ *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error) {
	if m.store == nil {
		m.store = map[uuid.UUID]entities.GenerationRequest{}
	}
	m.store[req.ID] = req
	m.updates++
	return req, nil
}

func (m *memReqRepo) single() entities.GenerationRequest {
	for _, v := range m.store {
		return v
	}
	return entities.GenerationRequest{}
}

type memScenRepo struct {
	rows []entities.GherkinScenario
}

func (m *memScenRepo) Create(_ context.Context, _ *gorm.DB, sc entities.GherkinScenario) (entities.GherkinScenario, error) {
	m.rows = append(m.rows, sc)
	return sc, nil
}

func validReq() dto.GenerateRequest {
	return dto.GenerateRequest{
		UserStory:    "As an admin I want to delete user data so that stale accounts are gone.",
		GuiElements:  []string{"Login button", "Password field"},
		FreeScenario: "Admin logs in, opens the page, clicks delete.",
	}
}

func validResp() dto.GenerateResponse {
	return dto.GenerateResponse{
		Feature: dto.Feature{
			FeatureName: "Delete User",
			Scenarios: []dto.Scenario{
				{
					Name: "Delete existing user", Kind: "positive", GuiElementIDs: []int{1},
					Given: []string{"the admin is on the user page with Login button visible"},
					When:  []string{"the admin clicks the Login button"},
					Then:  []string{"the dashboard appears"},
				},
				{
					Name: "Wrong password", Kind: "negative", GuiElementIDs: []int{2},
					Given: []string{"the admin is on the login page"},
					When:  []string{"the admin submits a wrong password in the Password field"},
					Then:  []string{"an error message appears"},
				},
			},
		},
		Gherkin: "Feature: Delete User\n  Scenario: Delete existing user\n    Given x\n",
	}
}

func newSvc(stub *stubGenerator, req *memReqRepo, sc *memScenRepo) service.ScenarioService {
	return service.NewScenarioService(stub, req, sc)
}

func TestGenerateInvalidUserID(t *testing.T) {
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		t.Fatal("generator tidak boleh dipanggil")
		return dto.GenerateResponse{}, nil
	}}
	reqRepo := &memReqRepo{}
	_, err := newSvc(stub, reqRepo, &memScenRepo{}).Generate(context.Background(), "bukan-uuid", validReq())
	if !errors.Is(err, service.ErrInvalidRequest) {
		t.Fatalf("error = %v, ingin ErrInvalidRequest", err)
	}
	if stub.calls != 0 || reqRepo.creates != 0 {
		t.Errorf("tidak boleh ada panggilan generator (%d) atau create (%d)", stub.calls, reqRepo.creates)
	}
}

func TestGenerateInvalidInput(t *testing.T) {
	base := validReq()
	tests := []struct {
		name   string
		change func(*dto.GenerateRequest)
	}{
		{"story terlalu pendek", func(r *dto.GenerateRequest) { r.UserStory = "short" }},
		{"tanpa elemen gui", func(r *dto.GenerateRequest) { r.GuiElements = nil }},
		{"free scenario kosong", func(r *dto.GenerateRequest) { r.FreeScenario = "   " }},
		{"elemen terlalu panjang", func(r *dto.GenerateRequest) { r.GuiElements = []string{strings.Repeat("x", 201)} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
				t.Fatal("generator tidak boleh dipanggil")
				return dto.GenerateResponse{}, nil
			}}
			reqRepo := &memReqRepo{}
			req := base
			tt.change(&req)
			_, err := newSvc(stub, reqRepo, &memScenRepo{}).Generate(context.Background(), uuid.NewString(), req)
			if !errors.Is(err, service.ErrInvalidRequest) {
				t.Fatalf("error = %v, ingin ErrInvalidRequest", err)
			}
			if stub.calls != 0 || reqRepo.creates != 0 {
				t.Error("request invalid tidak boleh menyentuh generator/DB")
			}
		})
	}
}

func TestGenerateSuccess(t *testing.T) {
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		return validResp(), nil
	}}
	reqRepo, scenRepo := &memReqRepo{}, &memScenRepo{}
	res, err := newSvc(stub, reqRepo, scenRepo).Generate(context.Background(), uuid.NewString(), validReq())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if stub.calls != 1 {
		t.Errorf("generator dipanggil %d kali, ingin 1", stub.calls)
	}
	saved := reqRepo.single()
	if saved.Status != dto.RequestSuccess {
		t.Errorf("status request = %q, ingin SUCCESS", saved.Status)
	}
	if len(scenRepo.rows) != 1 {
		t.Fatalf("scenario tersimpan = %d, ingin 1", len(scenRepo.rows))
	}
	row := scenRepo.rows[0]
	if row.Status != dto.ScenarioValid || row.SourceType != dto.SourceLLM {
		t.Errorf("row = %+v, ingin VALID/LLM", row)
	}
	if row.CoveredGuiCount != 2 || row.TotalGuiCount != 2 || row.CoveragePercent != 100 {
		t.Errorf("coverage = %d/%d %.2f, ingin 2/2 100", row.CoveredGuiCount, row.TotalGuiCount, row.CoveragePercent)
	}
	if len(res.Feature.Warnings) != 0 {
		t.Errorf("warnings = %v, ingin kosong", res.Feature.Warnings)
	}
}

func TestGenerateUncoveredWarning(t *testing.T) {
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		return validResp(), nil
	}}
	reqRepo, scenRepo := &memReqRepo{}, &memScenRepo{}
	req := validReq()
	req.GuiElements = []string{"Login button", "Ghost element"}
	res, err := newSvc(stub, reqRepo, scenRepo).Generate(context.Background(), uuid.NewString(), req)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	row := scenRepo.rows[0]
	if row.CoveredGuiCount != 1 || row.CoveragePercent != 50 {
		t.Errorf("coverage = %d %.2f, ingin 1 50", row.CoveredGuiCount, row.CoveragePercent)
	}
	found := false
	for _, w := range res.Feature.Warnings {
		if strings.Contains(w, "Ghost element") {
			found = true
		}
	}
	if !found {
		t.Errorf("warning elemen tak tercakup hilang: %v", res.Feature.Warnings)
	}
}

func TestGenerateInvalidResponseFails(t *testing.T) {
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		return dto.GenerateResponse{Gherkin: "  ", Feature: dto.Feature{FeatureName: "X"}}, nil
	}}
	reqRepo, scenRepo := &memReqRepo{}, &memScenRepo{}
	_, err := newSvc(stub, reqRepo, scenRepo).Generate(context.Background(), uuid.NewString(), validReq())
	if !errors.Is(err, service.ErrBadGateway) {
		t.Fatalf("error = %v, ingin ErrBadGateway", err)
	}
	if stub.calls != service.MaxAttempts {
		t.Errorf("retry = %d kali, ingin %d", stub.calls, service.MaxAttempts)
	}
	saved := reqRepo.single()
	if saved.Status != dto.RequestFailed || saved.ErrorMessage == "" {
		t.Errorf("request harus FAILED + error_message: %+v", saved)
	}
	if len(scenRepo.rows) != 1 || scenRepo.rows[0].Status != dto.ScenarioInvalid {
		t.Errorf("hasil invalid harus tersimpan sebagai INVALID: %+v", scenRepo.rows)
	}
}

func TestGenerateTransientRetrySucceeds(t *testing.T) {
	n := 0
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		n++
		if n < 3 {
			return dto.GenerateResponse{}, service.ErrUnavailable
		}
		return validResp(), nil
	}}
	reqRepo, scenRepo := &memReqRepo{}, &memScenRepo{}
	_, err := newSvc(stub, reqRepo, scenRepo).Generate(context.Background(), uuid.NewString(), validReq())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if stub.calls != 3 {
		t.Errorf("generator dipanggil %d kali, ingin 3", stub.calls)
	}
	if reqRepo.single().Status != dto.RequestSuccess {
		t.Error("request harus SUCCESS setelah retry berhasil")
	}
}

func TestGenerateTimeoutFails(t *testing.T) {
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		return dto.GenerateResponse{}, service.ErrTimeout
	}}
	reqRepo, scenRepo := &memReqRepo{}, &memScenRepo{}
	_, err := newSvc(stub, reqRepo, scenRepo).Generate(context.Background(), uuid.NewString(), validReq())
	if !errors.Is(err, service.ErrTimeout) {
		t.Fatalf("error = %v, ingin ErrTimeout", err)
	}
	if stub.calls != service.MaxAttempts {
		t.Errorf("retry = %d kali, ingin %d", stub.calls, service.MaxAttempts)
	}
	if len(scenRepo.rows) != 0 {
		t.Error("timeout tanpa respons tidak boleh menyimpan scenario")
	}
	if reqRepo.single().Status != dto.RequestFailed {
		t.Error("request harus FAILED")
	}
}

func TestGenerateIdenticalOutputStopsEarly(t *testing.T) {
	bad := dto.GenerateResponse{Gherkin: "Feature: X\n", Feature: dto.Feature{FeatureName: "X"}}
	stub := &stubGenerator{fn: func(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
		return bad, nil
	}}
	reqRepo := &memReqRepo{}
	_, err := newSvc(stub, reqRepo, &memScenRepo{}).Generate(context.Background(), uuid.NewString(), validReq())
	if !errors.Is(err, service.ErrBadGateway) {
		t.Fatalf("error = %v, ingin ErrBadGateway", err)
	}
	if stub.calls != 2 {
		t.Errorf("output identik harus stop di 2 panggilan, dapat %d", stub.calls)
	}
}
