package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authDto "github.com/story2gherkin/modules/auth/dto"
	"github.com/story2gherkin/pkg/helpers"
	"github.com/story2gherkin/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const passwordResetLifetime = 15 * time.Minute

func passwordResetHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newPasswordResetToken() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf(
			"generate password reset token: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func validPasswordResetToken(token string) bool {
	if len(token) != 43 {
		return false
	}

	decoded, err := base64.RawURLEncoding.Strict().
		DecodeString(token)

	return err == nil && len(decoded) == 32
}

func validNewPassword(password string) bool {
	if !utf8.ValidString(password) ||
		utf8.RuneCountInString(password) < 8 ||
		len(password) > 72 {
		return false
	}

	hasUpper := false
	hasLower := false

	for _, character := range password {
		if unicode.IsSpace(character) || character == '\uFEFF' {
			return false
		}

		if character >= 'A' && character <= 'Z' {
			hasUpper = true
		}

		if character >= 'a' && character <= 'z' {
			hasLower = true
		}
	}

	return hasUpper && hasLower
}

func passwordResetURL(token string) (string, error) {
	frontendURL := strings.TrimSpace(
		os.Getenv("FRONTEND_URL"),
	)

	if frontendURL == "" {
		return "", errors.New("FRONTEND_URL is required")
	}

	parsed, err := url.Parse(frontendURL)
	if err != nil {
		return "", errors.New("FRONTEND_URL is invalid")
	}

	if parsed.Hostname() == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.ForceQuery ||
		parsed.Fragment != "" {
		return "", errors.New("FRONTEND_URL is invalid")
	}

	localHost := parsed.Hostname() == "localhost" ||
		parsed.Hostname() == "127.0.0.1" ||
		parsed.Hostname() == "::1"

	if parsed.Scheme != "https" &&
		!(parsed.Scheme == "http" && localHost) {
		return "", errors.New(
			"FRONTEND_URL must use HTTPS outside localhost",
		)
	}

	parsed.Path = strings.TrimRight(
		parsed.Path,
		"/",
	) + "/reset-password"

	parsed.RawPath = ""

	query := url.Values{}
	query.Set("token", token)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func (s *authService) SendPasswordReset(
	ctx context.Context,
	req authDto.SendPasswordResetRequest,
) error {
	email := strings.ToLower(
		strings.TrimSpace(req.Email),
	)

	user, err := s.userRepository.GetUserByEmail(
		ctx,
		s.db,
		email,
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Keep the successful response the same for unknown emails.
		return nil
	}

	if err != nil {
		return fmt.Errorf(
			"look up password reset account: %w",
			err,
		)
	}

	if user.ID == uuid.Nil {
		return errors.New("password reset account has no ID")
	}

	if s.db == nil {
		return errors.New(
			"password reset database is not configured",
		)
	}

	token, err := newPasswordResetToken()
	if err != nil {
		return err
	}

	resetURL, err := passwordResetURL(token)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	record := entities.PasswordResetToken{
		UserID:    user.ID.String(),
		TokenHash: passwordResetHash(token),
		ExpiresAt: now.Add(passwordResetLifetime),
		CreatedAt: now,
	}

	// A new request replaces the previous token for this account.
	err = s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
			},
			DoUpdates: clause.AssignmentColumns(
				[]string{
					"token_hash",
					"expires_at",
					"created_at",
				},
			),
		}).
		Create(&record).
		Error

	if err != nil {
		return fmt.Errorf(
			"store password reset token: %w",
			err,
		)
	}

	body := `<p>We received a request to reset your Story2Gherkin password.</p>
<p><a href="` + html.EscapeString(resetURL) + `">Reset Password</a></p>
<p>This link expires in 15 minutes and can only be used once.</p>
<p>If you did not request this change, you can ignore this email.</p>`

	if err := utils.SendMail(
		user.Email,
		"Reset your Story2Gherkin password",
		body,
	); err != nil {
		// Do not remove a newer token created by another request.
		s.db.WithContext(ctx).
			Where("token_hash = ?", record.TokenHash).
			Delete(&entities.PasswordResetToken{})

		return fmt.Errorf(
			"send password reset email: %w",
			err,
		)
	}

	return nil
}

func (s *authService) ResetPassword(
	ctx context.Context,
	req authDto.ResetPasswordRequest,
) error {
	token := strings.TrimSpace(req.Token)

	if !validPasswordResetToken(token) {
		return authDto.ErrPasswordResetToken
	}

	if !validNewPassword(req.NewPassword) {
		return authDto.ErrInvalidNewPassword
	}

	if s.db == nil {
		return errors.New(
			"password reset database is not configured",
		)
	}

	return s.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var record entities.PasswordResetToken

			// DELETE ... RETURNING claims the token atomically.
			// Concurrent requests cannot both consume the same token.
			result := tx.
				Clauses(clause.Returning{
					Columns: []clause.Column{
						{Name: "user_id"},
					},
				}).
				Where(
					"token_hash = ? AND expires_at > CURRENT_TIMESTAMP",
					passwordResetHash(token),
				).
				Delete(&record)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected != 1 {
				return authDto.ErrPasswordResetToken
			}

			hashedPassword, err := helpers.HashPassword(
				req.NewPassword,
			)

			if err != nil {
				return err
			}

			result = tx.Model(&entities.User{}).
				Where("id = ?", record.UserID).
				Update("password", hashedPassword)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected != 1 {
				return authDto.ErrPasswordResetToken
			}

			// Revoke stored refresh tokens for this account.
			return tx.
				Where("user_id = ?", record.UserID).
				Delete(&entities.RefreshToken{}).
				Error
		},
	)
}
