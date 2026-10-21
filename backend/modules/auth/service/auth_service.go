package service

import (
	"context"
	"errors"
	"fmt"
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
}

type authService struct {
	userRepository repository.UserRepository
	jwtService     JWTService
	db             *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository: userRepo,
		jwtService:     jwtService,
		db:            db,
	}
}

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

func (s *authService) Login(
	ctx context.Context,
	req authDto.LoginRequest,
) (authDto.LoginResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Password tidak di-trim karena spasi bisa menjadi bagian password.
	if req.Email == "" || req.Password == "" || len(req.Password) > 72 {
		return authDto.LoginResponse{}, ErrInvalidCredentials
	}

	// 1. Cari pengguna berdasarkan email.
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

	// 2. Bandingkan password dengan hash yang tersimpan.
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

	// 3. Buat JWT menggunakan identitas dari database.
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

	// 4. Kembalikan token beserta informasi pengguna.
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