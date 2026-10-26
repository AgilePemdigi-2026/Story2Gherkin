package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"gorm.io/gorm"
)

func TestSendPasswordReset_UserNotFound_ReturnsNil(t *testing.T) {
	repo := &fakeUserRepository{
		getByEmailFn: func(context.Context, string) (entities.User, error) {
			return entities.User{}, gorm.ErrRecordNotFound
		},
	}
	svc := service.NewAuthService(repo, nil, nil, nil)
	if err := svc.SendPasswordReset(context.Background(), dto.SendPasswordResetRequest{Email: "hilang@example.com"}); err != nil {
		t.Fatalf("SendPasswordReset() error = %v, want nil (anti-enumerasi)", err)
	}
}

func TestSendPasswordReset_DBError_ReturnsEmailNotFound(t *testing.T) {
	dbErr := errors.New("db down")
	repo := &fakeUserRepository{
		getByEmailFn: func(context.Context, string) (entities.User, error) {
			return entities.User{}, dbErr
		},
	}
	svc := service.NewAuthService(repo, nil, nil, nil)
	if err := svc.SendPasswordReset(context.Background(), dto.SendPasswordResetRequest{Email: "a@example.com"}); !errors.Is(err, userDto.ErrEmailNotFound) {
		t.Fatalf("SendPasswordReset() error = %v, want ErrEmailNotFound", err)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := service.NewAuthService(repo, service.NewJWTService(), nil, nil)
	err := svc.ResetPassword(context.Background(), dto.ResetPasswordRequest{Token: "token-bukan-jwt", NewPassword: "rahasia123"})
	if !errors.Is(err, dto.ErrPasswordResetToken) {
		t.Fatalf("ResetPassword() error = %v, want ErrPasswordResetToken", err)
	}
}

func TestResetPassword_ValidToken_UserNotFound(t *testing.T) {
	jwt := service.NewJWTService()
	token := jwt.GenerateAccessToken("00000000-0000-4000-8000-000000000001", "password_reset")
	repo := &fakeUserRepository{
		getByIDFn: func(context.Context, string) (entities.User, error) {
			return entities.User{}, gorm.ErrRecordNotFound
		},
	}
	svc := service.NewAuthService(repo, jwt, nil, nil)
	err := svc.ResetPassword(context.Background(), dto.ResetPasswordRequest{Token: token, NewPassword: "rahasia123"})
	if !errors.Is(err, userDto.ErrUserNotFound) {
		t.Fatalf("ResetPassword() error = %v, want ErrUserNotFound", err)
	}
}
