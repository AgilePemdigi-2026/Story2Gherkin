package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

// hashResetToken mengubah token mentah menjadi SHA-256 hash.
func hashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateResetToken menghasilkan token acak yang aman
// untuk dikirim melalui link reset password.
func generateResetToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// createPasswordResetToken membuat dan menyimpan reset token.
// Nilai token mentah hanya dikembalikan ke service internal
// untuk dikirim melalui email.
func createPasswordResetToken(
	ctx context.Context,
	db *gorm.DB,
	userID uuid.UUID,
) (string, error) {

	token, err := generateResetToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()

	resetToken := entities.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashResetToken(token),
		ExpiresAt: now.Add(30 * time.Minute),
	}

	// Token reset lama untuk akun yang sama dicabut.
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.PasswordResetToken{}).
			Where("user_id = ? AND used_at IS NULL", userID).
			Update("used_at", now).Error; err != nil {
			return err
		}

		return tx.Create(&resetToken).Error
	})

	if err != nil {
		return "", err
	}

	return token, nil
}
