package service

import (
	"context"

	"github.com/story2gherkin/database/entities"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error)
	// Login, Logout, Forgot Password
}

type authService struct {
	userRepository         repository.UserRepository
	jwtService             JWTService
	db                     *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepository:         userRepo,
		jwtService:             jwtService,
		db:                     db,
	}
}

func (s *authService) Register(ctx context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
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
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		Role:       entities.Role(req.Role),
	}

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
	if err != nil {
		return userDto.UserResponse{}, err
	}

	return userDto.UserResponse{
		ID:         createdUser.ID.String(),
		Name:       createdUser.Name,
		Email:      createdUser.Email,
		Role:       string(createdUser.Role),
	}, nil
}