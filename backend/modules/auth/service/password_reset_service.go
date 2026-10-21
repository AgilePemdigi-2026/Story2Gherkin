package service

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/pkg/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidResetToken = errors.New(
		"invalid, expired, or already used reset token",
	)

	ErrInvalidNewPassword = errors.New(
		"password must be at least 8 characters and at most 72 bytes",
	)
)

func (s *authService) ResetPassword(
	ctx context.Context,
	token string,
	newPassword string,
) error {

	// 1. Validasi password dan token.
	if token == "" {
		return ErrInvalidResetToken
	}

	if utf8.RuneCountInString(newPassword) < 8 ||
		len([]byte(newPassword)) > 72 {
		return ErrInvalidNewPassword
	}

	// 2. Hash password baru menggunakan bcrypt.
	hashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 3. Hitung hash token dari link email.
	tokenHash := hashResetToken(token)

	// 4. Update password dan token secara atomik.
	return s.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var resetToken entities.PasswordResetToken

			now := time.Now().UTC()

			// Kunci record agar token tidak digunakan
			// dua kali secara bersamaan.
			err := tx.Clauses(
				clause.Locking{Strength: "UPDATE"},
			).
				Where(
					"token_hash = ? AND used_at IS NULL AND expires_at > ?",
					tokenHash,
					now,
				).
				Take(&resetToken).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidResetToken
			}
			if err != nil {
				return err
			}

			// 5. Update password akun pemilik token.
			result := tx.Model(&entities.User{}).
				Where("id = ?", resetToken.UserID.String()).
				Update("password", hashedPassword)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected != 1 {
				return ErrInvalidResetToken
			}

			// 6. Tandai token sudah digunakan.
			result = tx.Model(&entities.PasswordResetToken{}).
				Where("id = ? AND used_at IS NULL", resetToken.ID).
				Update("used_at", now)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected != 1 {
				return ErrInvalidResetToken
			}

			return nil
		},
	)
}
