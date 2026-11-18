package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/modules/auth/service"
	"gorm.io/gorm"
)

func TestSendPasswordReset_UserNotFound_ReturnsNil(t *testing.T) {
	repo := &fakeUserRepository{
		getByEmailFn: func(
			context.Context,
			string,
		) (entities.User, error) {
			return entities.User{}, gorm.ErrRecordNotFound
		},
	}

	svc := service.NewAuthService(repo, nil, nil, nil)

	err := svc.SendPasswordReset(
		context.Background(),
		dto.SendPasswordResetRequest{
			Email: "missing@example.com",
		},
	)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
}

func TestSendPasswordReset_DBError_IsPreserved(t *testing.T) {
	dbErr := errors.New("database unavailable")

	repo := &fakeUserRepository{
		getByEmailFn: func(
			context.Context,
			string,
		) (entities.User, error) {
			return entities.User{}, dbErr
		},
	}

	svc := service.NewAuthService(repo, nil, nil, nil)

	err := svc.SendPasswordReset(
		context.Background(),
		dto.SendPasswordResetRequest{
			Email: "user@example.com",
		},
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf("error = %v, want database error", err)
	}
}

func TestResetPassword_RejectsInvalidTokens(t *testing.T) {
	svc := service.NewAuthService(nil, nil, nil, nil)

	tokens := []string{
		"",
		"invalid-token",
		strings.Repeat("A", 42),
		strings.Repeat("!", 43),
	}

	for _, token := range tokens {
		err := svc.ResetPassword(
			context.Background(),
			dto.ResetPasswordRequest{
				Token:       token,
				NewPassword: "NewPassword123",
			},
		)

		if !errors.Is(err, dto.ErrPasswordResetToken) {
			t.Fatalf(
				"token %q: error = %v, want invalid token",
				token,
				err,
			)
		}
	}
}

func TestResetPassword_RejectsJWT(t *testing.T) {
	jwtService := service.NewJWTService()

	for _, role := range []string{"qa", "password_reset"} {
		token := jwtService.GenerateAccessToken(
			"00000000-0000-4000-8000-000000000001",
			role,
		)

		svc := service.NewAuthService(
			nil,
			jwtService,
			nil,
			nil,
		)

		err := svc.ResetPassword(
			context.Background(),
			dto.ResetPasswordRequest{
				Token:       token,
				NewPassword: "NewPassword123",
			},
		)

		if !errors.Is(err, dto.ErrPasswordResetToken) {
			t.Fatalf(
				"role %q: error = %v, want invalid reset token",
				role,
				err,
			)
		}
	}
}

func TestResetPassword_RejectsInvalidPasswords(t *testing.T) {
	svc := service.NewAuthService(nil, nil, nil, nil)

	// A valid encoding of 32 bytes; no database access is expected.
	token := strings.Repeat("A", 43)

	passwords := []string{
		"Short1",
		"lowercase123",
		"UPPERCASE123",
		"A" + strings.Repeat("a", 72),
	}

	for _, password := range passwords {
		err := svc.ResetPassword(
			context.Background(),
			dto.ResetPasswordRequest{
				Token:       token,
				NewPassword: password,
			},
		)

		if !errors.Is(err, dto.ErrInvalidNewPassword) {
			t.Fatalf(
				"error = %v, want invalid new password",
				err,
			)
		}
	}
}
