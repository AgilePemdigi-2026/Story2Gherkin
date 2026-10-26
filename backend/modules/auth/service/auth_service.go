package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authDto "github.com/story2gherkin/modules/auth/dto"
	authRepo "github.com/story2gherkin/modules/auth/repository"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/helpers"
	"github.com/story2gherkin/pkg/utils"
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
		req userDto.UserLoginRequest,
	) (authDto.TokenResponse, error)

	RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (authDto.TokenResponse, error)
	SendPasswordReset(ctx context.Context, req authDto.SendPasswordResetRequest) error
	ResetPassword(ctx context.Context, req authDto.ResetPasswordRequest) error
}

type authService struct {
	userRepository         repository.UserRepository
	jwtService             JWTService
	refreshTokenRepository authRepo.RefreshTokenRepository
	db                     *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService JWTService,
	db *gorm.DB,
	refreshTokenRepo authRepo.RefreshTokenRepository,
) AuthService {
	return &authService{
		userRepository:         userRepo,
		refreshTokenRepository: refreshTokenRepo,
		jwtService:             jwtService,
		db:                     db,
	}
}

func (s *authService) Register(
	ctx context.Context,
	req userDto.UserCreateRequest,
) (userDto.UserResponse, error) {

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
	req userDto.UserLoginRequest,
) (authDto.TokenResponse, error) {

	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return authDto.TokenResponse{}, userDto.ErrEmailNotFound
	}

	isValid, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !isValid {
		return authDto.TokenResponse{}, authDto.ErrInvalidCredentials
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), string(user.Role))
	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	refreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, refreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	return authDto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         string(user.Role),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req authDto.RefreshTokenRequest) (authDto.TokenResponse, error) {
	refreshToken, err := s.refreshTokenRepository.FindByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, authDto.ErrRefreshTokenNotFound
	}

	accessToken := s.jwtService.GenerateAccessToken(refreshToken.UserID.String(), string(refreshToken.User.Role))
	newRefreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	err = s.refreshTokenRepository.DeleteByToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	newRefreshToken := entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: expiresAt,
	}

	_, err = s.refreshTokenRepository.Create(ctx, s.db, newRefreshToken)
	if err != nil {
		return authDto.TokenResponse{}, err
	}

	return authDto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		Role:         string(refreshToken.User.Role),
	}, nil
}

func (s *authService) SendPasswordReset(ctx context.Context, req authDto.SendPasswordResetRequest) error {
	user, err := s.userRepository.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return userDto.ErrEmailNotFound
	}

	resetToken := s.jwtService.GenerateAccessToken(user.ID.String(), "password_reset")

	subject := "Password Reset"
	body := "Please reset your password using this token: " + resetToken

	return utils.SendMail(user.Email, subject, body)
}

func (s *authService) ResetPassword(ctx context.Context, req authDto.ResetPasswordRequest) error {
	token, err := s.jwtService.ValidateToken(req.Token)
	if err != nil || !token.Valid {
		return authDto.ErrPasswordResetToken
	}

	userId, err := s.jwtService.GetUserIDByToken(req.Token)
	if err != nil {
		return authDto.ErrPasswordResetToken
	}

	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return userDto.ErrUserNotFound
	}

	hashedPassword, err := helpers.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	_, err = s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return err
	}

	return nil
}
