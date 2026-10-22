package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	authDto "github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/modules/auth/handler"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/pkg/constants"
	"gorm.io/gorm"
)

type fakeAuthService struct {
	registerFn func(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	gotReq     *userDto.UserCreateRequest
}

func (f *fakeAuthService) Register(
	ctx context.Context,
	req userDto.UserCreateRequest,
) (userDto.UserResponse, error) {
	f.gotReq = &req
	return f.registerFn(ctx, req)
}

func (f *fakeAuthService) Login(context.Context, authDto.LoginRequest) (authDto.LoginResponse, error) {
	return authDto.LoginResponse{}, nil
}

func (f *fakeAuthService) ForgotPassword(context.Context, string) error { return nil }

func (f *fakeAuthService) ResetPassword(context.Context, string, string) error { return nil }

// newTestHandler membangun handler lewat konstruktor aslinya.
// DB tidak dipakai oleh Register, cukup injector dengan *gorm.DB kosong.
func newTestHandler(t *testing.T, as service.AuthService) handler.AuthHandler {
	t.Helper()

	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(*do.Injector) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})
	return handler.NewAuthController(injector, as)
}

func doRegister(t *testing.T, h handler.AuthHandler, body string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("gagal decode response: %v, body=%s", err, rec.Body.String())
	}
	return res
}

func TestRegisterHandlerSuccess(t *testing.T) {
	as := &fakeAuthService{
		registerFn: func(_ context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
			return userDto.UserResponse{
				ID:    "user-1",
				Name:  req.Name,
				Email: req.Email,
				Role:  req.Role,
			}, nil
		},
	}
	h := newTestHandler(t, as)

	rec := doRegister(t, h, `{
		"name": "Rizal",
		"email": "rizal@example.com",
		"password": "rahasia123",
		"role": "qa"
	}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	res := decodeResponse(t, rec)
	if res["status"] != true {
		t.Errorf("status response = %v, want true", res["status"])
	}
	if res["message"] != userDto.MESSAGE_SUCCESS_REGISTER_USER {
		t.Errorf("message = %v, want %q", res["message"], userDto.MESSAGE_SUCCESS_REGISTER_USER)
	}

	data, ok := res["data"].(map[string]any)
	if !ok {
		t.Fatalf("data bukan object: %v", res["data"])
	}
	if data["id"] != "user-1" || data["email"] != "rizal@example.com" {
		t.Errorf("data tidak sesuai: %+v", data)
	}

	if as.gotReq == nil || as.gotReq.Name != "Rizal" || as.gotReq.Password != "rahasia123" {
		t.Errorf("request tidak diteruskan ke service: %+v", as.gotReq)
	}
}

func TestRegisterHandlerInvalidBody(t *testing.T) {
	called := false
	as := &fakeAuthService{
		registerFn: func(context.Context, userDto.UserCreateRequest) (userDto.UserResponse, error) {
			called = true
			return userDto.UserResponse{}, nil
		},
	}
	h := newTestHandler(t, as)

	// Password kurang dari 8 karakter.
	rec := doRegister(t, h, `{
		"name": "Rizal",
		"email": "rizal@example.com",
		"password": "pendek",
		"role": "qa"
	}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if called {
		t.Error("service.Register tidak boleh dipanggil saat body invalid")
	}

	res := decodeResponse(t, rec)
	if res["status"] != false {
		t.Errorf("status response = %v, want false", res["status"])
	}
	if res["message"] != userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY {
		t.Errorf("message = %v, want %q", res["message"], userDto.MESSAGE_FAILED_GET_DATA_FROM_BODY)
	}
}

func TestRegisterHandlerEmailAlreadyExists(t *testing.T) {
	as := &fakeAuthService{
		registerFn: func(context.Context, userDto.UserCreateRequest) (userDto.UserResponse, error) {
			return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
		},
	}
	h := newTestHandler(t, as)

	rec := doRegister(t, h, `{
		"name": "Rizal",
		"email": "rizal@example.com",
		"password": "rahasia123",
		"role": "qa"
	}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	res := decodeResponse(t, rec)
	if res["status"] != false {
		t.Errorf("status response = %v, want false", res["status"])
	}
	if res["error"] != "Email sudah terdaftar." {
		t.Errorf("error = %v, want %q", res["error"], "Email sudah terdaftar.")
	}
}

func TestRegisterHandlerServiceError(t *testing.T) {
	as := &fakeAuthService{
		registerFn: func(context.Context, userDto.UserCreateRequest) (userDto.UserResponse, error) {
			return userDto.UserResponse{}, errors.New("database error")
		},
	}
	h := newTestHandler(t, as)

	rec := doRegister(t, h, `{
		"name": "Rizal",
		"email": "rizal@example.com",
		"password": "rahasia123",
		"role": "qa"
	}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	res := decodeResponse(t, rec)
	if res["status"] != false {
		t.Errorf("status response = %v, want false", res["status"])
	}
	if res["error"] != "Terjadi kesalahan pada server." {
		t.Errorf("error = %v, want %q", res["error"], "Terjadi kesalahan pada server.")
	}
}
