package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
	tokenIssuer     = "story2gherkin"
)

type JWTService interface {
	GenerateAccessToken(userID, role string) (string, error)
	GenerateRefreshToken() (string, time.Time, error)
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDByToken(token string) (string, error)
	GetRoleByToken(token string) (string, error)
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey     string
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// DIUBAH: constructor mengembalikan error konfigurasi.
func NewJWTService() (JWTService, error) {
	secretKey := os.Getenv("JWT_SECRET")

	if strings.TrimSpace(secretKey) == "Template" ||
		len(strings.TrimSpace(secretKey)) < 32 {
		return nil, fmt.Errorf(
			"JWT_SECRET wajib diisi dengan secret acak minimal 32 byte",
		)
	}

	return &jwtService{
		secretKey:     secretKey,
		issuer:        tokenIssuer,
		accessExpiry:  AccessTokenTTL,
		refreshExpiry: RefreshTokenTTL,
	}, nil
}

func (j *jwtService) GenerateAccessToken(
	userID, role string,
) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user ID tidak boleh kosong")
	}

	if !isValidRole(role) {
		return "", fmt.Errorf("role tidak valid")
	}

	now := time.Now().UTC()

	claims := jwtCustomClaim{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("menandatangani access token: %w", err)
	}

	return signedToken, nil
}

// DIUBAH: error random generator diteruskan kepada pemanggil.
// Fungsi ini hanya membuat token; belum menyediakan alur refresh.
func (j *jwtService) GenerateRefreshToken() (string, time.Time, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", time.Time{}, fmt.Errorf(
			"membuat refresh token: %w",
			err,
		)
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	expiresAt := time.Now().UTC().Add(j.refreshExpiry)

	return refreshToken, expiresAt, nil
}

func (j *jwtService) ValidateToken(rawToken string) (*jwt.Token, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, fmt.Errorf("token tidak boleh kosong")
	}

	claims := &jwtCustomClaim{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(_ *jwt.Token) (interface{}, error) {
			return []byte(j.secretKey), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil {
		return nil, fmt.Errorf("validasi token gagal: %w", err)
	}

	if token == nil || !token.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}

	now := time.Now().UTC()

	if !claims.VerifyIssuer(j.issuer, true) {
		return nil, fmt.Errorf("issuer token tidak valid")
	}

	// Parameter true mewajibkan claim expiry tersedia.
	if !claims.VerifyExpiresAt(now, true) {
		return nil, fmt.Errorf(
			"token kedaluwarsa atau expiry tidak tersedia",
		)
	}

	// Wajib memiliki waktu penerbitan yang tidak berada di masa depan.
	if !claims.VerifyIssuedAt(now, true) {
		return nil, fmt.Errorf("waktu penerbitan token tidak valid")
	}

	if claims.UserID == "" || claims.Subject != claims.UserID {
		return nil, fmt.Errorf("identitas token tidak valid")
	}

	if !isValidRole(claims.Role) {
		return nil, fmt.Errorf("role token tidak valid")
	}

	return token, nil
}

func (j *jwtService) GetUserIDByToken(rawToken string) (string, error) {
	token, err := j.ValidateToken(rawToken)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwtCustomClaim)
	if !ok || claims == nil {
		return "", fmt.Errorf("claims token tidak valid")
	}

	return claims.UserID, nil
}

// DITAMBAHKAN ke interface agar dapat dipanggil melalui JWTService.
func (j *jwtService) GetRoleByToken(rawToken string) (string, error) {
	token, err := j.ValidateToken(rawToken)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwtCustomClaim)
	if !ok || claims == nil {
		return "", fmt.Errorf("claims token tidak valid")
	}

	return claims.Role, nil
}

func isValidRole(role string) bool {
	return role == "system_analyst" || role == "qa"
}