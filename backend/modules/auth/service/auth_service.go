package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authDto "github.com/story2gherkin/modules/auth/dto"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/helpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("email atau password salah")

type AuthService interface {
	Register(
		ctx context.Context,
		req userDto.UserCreateRequest,
	) (userDto.UserResponse, error)

	Login(
		ctx context.Context,
		req authDto.LoginRequest,
	) (authDto.LoginResponse, error)

	ForgotPassword(
		ctx context.Context,
		email string,
	) error

	ResetPassword(
		ctx context.Context,
		token string,
		newPassword string,
	) error
}

type authService struct {
	userRepository repository.UserRepository
	jwtService     JWTService
	mailer         PasswordResetMailer
	db             *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	db *gorm.DB,
	mailer PasswordResetMailer,
) AuthService {
	return &authService{
		userRepository: userRepo,
		jwtService:     jwtService,
		db:             db,
		mailer:         mailer,
	}
}

// Register membuat akun pengguna baru.
func (s *authService) Register(
	ctx context.Context,
	req userDto.UserCreateRequest,
) (userDto.UserResponse, error) {

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	_, isExist, err := s.userRepository.CheckEmail(
		ctx,
		s.db,
		req.Email,
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return userDto.UserResponse{}, err
	}

	if isExist {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	user := entities.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     entities.Role(req.Role),
	}

	createdUser, err := s.userRepository.Register(
		ctx,
		s.db,
		user,
	)

	if err != nil {
		return userDto.UserResponse{}, err
	}

	return userDto.UserResponse{
		ID:    createdUser.ID.String(),
		Name:  createdUser.Name,
		Email: createdUser.Email,
		Role:  string(createdUser.Role),
	}, nil
}

// Login memverifikasi kredensial dan menghasilkan JWT.
func (s *authService) Login(
	ctx context.Context,
	req authDto.LoginRequest,
) (authDto.LoginResponse, error) {

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Password tidak di-trim karena spasi
	// dapat menjadi bagian dari password.
	if req.Email == "" ||
		req.Password == "" ||
		len(req.Password) > 72 {
		return authDto.LoginResponse{}, ErrInvalidCredentials
	}

	// Cari pengguna berdasarkan email.
	user, isExist, err := s.userRepository.CheckEmail(
		ctx,
		s.db,
		req.Email,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return authDto.LoginResponse{}, ErrInvalidCredentials
		}

		return authDto.LoginResponse{}, fmt.Errorf(
			"mencari pengguna: %w",
			err,
		)
	}

	if !isExist {
		return authDto.LoginResponse{}, ErrInvalidCredentials
	}

	// Bandingkan password dengan hash tersimpan.
	matches, err := helpers.CheckPassword(
		user.Password,
		[]byte(req.Password),
	)

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return authDto.LoginResponse{}, ErrInvalidCredentials
		}

		return authDto.LoginResponse{}, fmt.Errorf(
			"memverifikasi password: %w",
			err,
		)
	}

	if !matches {
		return authDto.LoginResponse{}, ErrInvalidCredentials
	}

	// Buat JWT dari identitas pengguna.
	accessToken, err := s.jwtService.GenerateAccessToken(
		user.ID.String(),
		string(user.Role),
	)

	if err != nil {
		return authDto.LoginResponse{}, fmt.Errorf(
			"membuat access token: %w",
			err,
		)
	}

	// Kembalikan token dan informasi pengguna.
	return authDto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(AccessTokenTTL / time.Second),
		User: authDto.LoginUserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}, nil
}

// ForgotPassword memproses permintaan reset password.
func (s *authService) ForgotPassword(
	ctx context.Context,
	email string,
) error {

	email = strings.ToLower(strings.TrimSpace(email))

	user, exists, err := s.userRepository.CheckEmail(
		ctx,
		s.db,
		email,
	)

	// Jangan ungkapkan apakah email terdaftar.
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		(!exists && err == nil) {
		return nil
	}

	if err != nil {
		return err
	}

	frontendURL := strings.TrimRight(
		os.Getenv("FRONTEND_URL"),
		"/",
	)

	if frontendURL == "" {
		log.Println("Password reset: FRONTEND_URL is not configured")
		return nil
	}

	// Buat reset token dan simpan hash-nya.
	token, err := createPasswordResetToken(
		ctx,
		s.db,
		user.ID,
	)

	if err != nil {
		return err
	}

	resetURL := frontendURL +
		"/reset-password?token=" +
		url.QueryEscape(token)

	// Kirim tautan reset melalui email.
	if err := s.mailer.SendResetLink(
		user.Email,
		resetURL,
	); err != nil {
		log.Println("Password reset: email delivery failed")
		return nil
	}

	return nil
}
