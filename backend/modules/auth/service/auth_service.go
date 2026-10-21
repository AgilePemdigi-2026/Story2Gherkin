package service

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/helpers"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
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
		mailer:         mailer,
		db:             db,
	}
}

func (s *authService) Register(
	ctx context.Context,
	req userDto.UserCreateRequest,
) (userDto.UserResponse, error) {

	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
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

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
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

// ForgotPassword memproses permintaan reset password.
func (s *authService) ForgotPassword(
	ctx context.Context,
	email string,
) error {

	email = strings.TrimSpace(email)

	user, exists, err := s.userRepository.CheckEmail(ctx, s.db, email)

	// Jangan ungkapkan apakah email terdaftar.
	if errors.Is(err, gorm.ErrRecordNotFound) || !exists && err == nil {
		return nil
	}

	if err != nil {
		return err
	}

	frontendURL := strings.TrimRight(
		os.Getenv("FRONTEND_URL"), "/",
	)

	if frontendURL == "" {
		log.Println("Password reset: FRONTEND_URL is not configured")
		return nil
	}

	// Membuat token dan menyimpan hash-nya.
	token, err := createPasswordResetToken(ctx, s.db, user.ID)
	if err != nil {
		return err
	}

	resetURL := frontendURL +
		"/reset-password?token=" +
		url.QueryEscape(token)

	// Mengirim tautan ke alamat email pengguna.
	if err := s.mailer.SendResetLink(user.Email, resetURL); err != nil {
		// Jangan mengungkapkan keberadaan akun lewat respons.
		log.Println("Password reset: email delivery failed")
		return nil
	}

	return nil
}
