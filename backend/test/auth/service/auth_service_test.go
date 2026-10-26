package service_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authDto "github.com/story2gherkin/modules/auth/dto"
	authRepo "github.com/story2gherkin/modules/auth/repository"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
	"testing"
	"time"
)

type fakeUserRepository struct {
	checkEmailFn func(ctx context.Context, email string) (entities.User, bool, error)
	registerFn   func(ctx context.Context, user entities.User) (entities.User, error)
	getByEmailFn func(ctx context.Context, email string) (entities.User, error)
	getByIDFn    func(ctx context.Context, id string) (entities.User, error)
	updateFn     func(ctx context.Context, user entities.User) (entities.User, error)

	checkedEmail string
	registered   *entities.User
	registerCall int
}

func (f *fakeUserRepository) CheckEmail(
	ctx context.Context,
	_ *gorm.DB,
	email string,
) (entities.User, bool, error) {
	f.checkedEmail = email
	return f.checkEmailFn(ctx, email)
}

func (f *fakeUserRepository) Register(
	ctx context.Context,
	_ *gorm.DB,
	user entities.User,
) (entities.User, error) {
	f.registerCall++
	f.registered = &user
	return f.registerFn(ctx, user)
}

func (f *fakeUserRepository) GetUserByEmail(ctx context.Context, _ *gorm.DB, email string) (entities.User, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}
	if f.checkEmailFn != nil {
		u, ok, err := f.checkEmailFn(ctx, email)
		if err != nil {
			return entities.User{}, err
		}
		if !ok {
			return entities.User{}, gorm.ErrRecordNotFound
		}
		return u, nil
	}
	return entities.User{}, gorm.ErrRecordNotFound
}

func (f *fakeUserRepository) GetUserById(ctx context.Context, _ *gorm.DB, id string) (entities.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return entities.User{}, gorm.ErrRecordNotFound
}

func (f *fakeUserRepository) Update(ctx context.Context, _ *gorm.DB, user entities.User) (entities.User, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, user)
	}
	return user, nil
}

func (f *fakeUserRepository) Delete(context.Context, *gorm.DB, string) error { return nil }

func newTestService(repo *fakeUserRepository) service.AuthService {
	return service.NewAuthService(repo, nil, nil, nil)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &fakeUserRepository{
		checkEmailFn: func(context.Context, string) (entities.User, bool, error) {
			return entities.User{}, false, gorm.ErrRecordNotFound
		},
		registerFn: func(_ context.Context, user entities.User) (entities.User, error) {
			return user, nil
		},
	}

	result, err := newTestService(repo).Register(context.Background(), userDto.UserCreateRequest{
		Name:     "  Rizal  ",
		Email:    "  RIZAL@Example.COM ",
		Password: "rahasia123",
		Role:     string(entities.RoleQA),
	})
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	if repo.checkedEmail != "rizal@example.com" {
		t.Errorf("CheckEmail() email = %q, want %q", repo.checkedEmail, "rizal@example.com")
	}
	if repo.registered == nil {
		t.Fatal("Register() tidak memanggil repository.Register")
	}
	if repo.registered.Name != "Rizal" {
		t.Errorf("nama tersimpan = %q, want %q", repo.registered.Name, "Rizal")
	}
	if repo.registered.Email != "rizal@example.com" {
		t.Errorf("email tersimpan = %q, want %q", repo.registered.Email, "rizal@example.com")
	}
	if repo.registered.Password == "rahasia123" {
		t.Error("password tersimpan masih plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.registered.Password), []byte("rahasia123")); err != nil {
		t.Errorf("password hash tidak cocok: %v", err)
	}
	if repo.registered.Role != entities.RoleQA {
		t.Errorf("role tersimpan = %q, want %q", repo.registered.Role, entities.RoleQA)
	}

	if result.ID == "" {
		t.Error("response ID kosong")
	}
	if result.Name != "Rizal" || result.Email != "rizal@example.com" || result.Role != string(entities.RoleQA) {
		t.Errorf("response tidak sesuai: %+v", result)
	}
}

func TestRegisterEmailAlreadyExists(t *testing.T) {
	repo := &fakeUserRepository{
		checkEmailFn: func(context.Context, string) (entities.User, bool, error) {
			return entities.User{}, true, nil
		},
		registerFn: func(context.Context, entities.User) (entities.User, error) {
			t.Fatal("repository.Register tidak boleh dipanggil")
			return entities.User{}, nil
		},
	}

	_, err := newTestService(repo).Register(context.Background(), userDto.UserCreateRequest{
		Name:     "Rizal",
		Email:    "rizal@example.com",
		Password: "rahasia123",
		Role:     string(entities.RoleQA),
	})

	if !errors.Is(err, userDto.ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
	if repo.registerCall != 0 {
		t.Errorf("repository.Register dipanggil %d kali, want 0", repo.registerCall)
	}
}

func TestRegisterCheckEmailError(t *testing.T) {
	dbErr := errors.New("koneksi database putus")
	repo := &fakeUserRepository{
		checkEmailFn: func(context.Context, string) (entities.User, bool, error) {
			return entities.User{}, false, dbErr
		},
	}

	_, err := newTestService(repo).Register(context.Background(), userDto.UserCreateRequest{
		Name:     "Rizal",
		Email:    "rizal@example.com",
		Password: "rahasia123",
		Role:     string(entities.RoleQA),
	})

	if !errors.Is(err, dbErr) {
		t.Fatalf("Register() error = %v, want %v", err, dbErr)
	}
}

func TestRegisterCreateError(t *testing.T) {
	createErr := errors.New("insert gagal")
	repo := &fakeUserRepository{
		checkEmailFn: func(context.Context, string) (entities.User, bool, error) {
			return entities.User{}, false, gorm.ErrRecordNotFound
		},
		registerFn: func(context.Context, entities.User) (entities.User, error) {
			return entities.User{}, createErr
		},
	}

	_, err := newTestService(repo).Register(context.Background(), userDto.UserCreateRequest{
		Name:     "Rizal",
		Email:    "rizal@example.com",
		Password: "rahasia123",
		Role:     string(entities.RoleQA),
	})

	if !errors.Is(err, createErr) {
		t.Fatalf("Register() error = %v, want %v", err, createErr)
	}
}

func TestRegisterHashError(t *testing.T) {
	repo := &fakeUserRepository{
		checkEmailFn: func(context.Context, string) (entities.User, bool, error) {
			return entities.User{}, false, gorm.ErrRecordNotFound
		},
		registerFn: func(context.Context, entities.User) (entities.User, error) {
			t.Fatal("repository.Register tidak boleh dipanggil")
			return entities.User{}, nil
		},
	}

	// bcrypt menolak password lebih dari 72 byte.
	_, err := newTestService(repo).Register(context.Background(), userDto.UserCreateRequest{
		Name:     "Rizal",
		Email:    "rizal@example.com",
		Password: strings.Repeat("a", 73),
		Role:     string(entities.RoleQA),
	})

	if err == nil {
		t.Fatal("Register() error = nil, want error hash password")
	}
	if repo.registerCall != 0 {
		t.Errorf("repository.Register dipanggil %d kali, want 0", repo.registerCall)
	}
}

type loginJWTStub struct {
	service.JWTService

	generate func(string, string) string
}

func (j *loginJWTStub) GenerateAccessToken(
	userID string,
	role string,
) string {
	return j.generate(userID, role)
}

func (j *loginJWTStub) GenerateRefreshToken() (string, time.Time) {
	return "refresh-uji", time.Now().Add(time.Hour)
}

type stubRefreshRepo struct {
	authRepo.RefreshTokenRepository
	createFn func(ctx context.Context, token entities.RefreshToken) (entities.RefreshToken, error)
}

func (s *stubRefreshRepo) Create(ctx context.Context, _ *gorm.DB, token entities.RefreshToken) (entities.RefreshToken, error) {
	if s.createFn != nil {
		return s.createFn(ctx, token)
	}
	return token, nil
}

type loginTestContextKey struct{}

func TestLoginService(t *testing.T) {
	// Spasi sengaja digunakan untuk memastikan password tidak di-trim.
	password := " PasswordUji123! "

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.MinCost,
	)
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.MustParse(
		"00000000-0000-4000-8000-000000000001",
	)

	dbErr := errors.New("database sedang bermasalah")

	validRequest := userDto.UserLoginRequest{
		Email:    "  ORWIN@EXAMPLE.COM  ",
		Password: password,
	}

	tests := []struct {
		name          string
		request       *userDto.UserLoginRequest
		role          entities.Role
		missingUser   bool
		repositoryErr error
		invalidHash   bool
		wantErr       error
		wantRepoCalls int
		wantJWTCalls  int
	}{
		{
			name:          "berhasil dan email dinormalisasi",
			wantRepoCalls: 1,
			wantJWTCalls:  1,
		},
		{
			name:          "role QA diteruskan dari database",
			role:          entities.RoleQA,
			wantRepoCalls: 1,
			wantJWTCalls:  1,
		},
		{
			name:          "pengguna tidak ditemukan",
			repositoryErr: gorm.ErrRecordNotFound,
			wantErr:       service.ErrInvalidCredentials,
			wantRepoCalls: 1,
		},
		{
			name: "record not found dibungkus",
			repositoryErr: fmt.Errorf(
				"repository: %w",
				gorm.ErrRecordNotFound,
			),
			wantErr:       service.ErrInvalidCredentials,
			wantRepoCalls: 1,
		},
		{
			name:          "pengguna tidak tersedia tanpa error",
			missingUser:   true,
			wantErr:       service.ErrInvalidCredentials,
			wantRepoCalls: 1,
		},
		{
			name:          "error database diteruskan",
			repositoryErr: dbErr,
			wantErr:       dbErr,
			wantRepoCalls: 1,
		},
		{
			name: "password salah",
			request: &userDto.UserLoginRequest{
				Email:    "orwin@example.com",
				Password: "PasswordSalah123!",
			},
			wantErr:       service.ErrInvalidCredentials,
			wantRepoCalls: 1,
		},
		{
			name:          "hash password rusak",
			invalidHash:   true,
			wantErr:       bcrypt.ErrHashTooShort,
			wantRepoCalls: 1,
		},
		{
			name: "email kosong",
			request: &userDto.UserLoginRequest{
				Email:    "   ",
				Password: password,
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "password kosong",
			request: &userDto.UserLoginRequest{
				Email:    "orwin@example.com",
				Password: "",
			},
			wantErr: service.ErrInvalidCredentials,
		},
		{
			name: "password UTF8 melebihi 72 byte",
			request: &userDto.UserLoginRequest{
				Email:    "orwin@example.com",
				Password: strings.Repeat("界", 25),
			},
			wantErr: service.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := entities.User{
				ID:       userID,
				Name:     "Orwin",
				Email:    "orwin@example.com",
				Password: string(hash),
				Role:     entities.RoleSystemAnalyst,
			}

			if tt.role != "" {
				user.Role = tt.role
			}

			if tt.invalidHash {
				user.Password = "hash-rusak"
			}

			req := validRequest
			if tt.request != nil {
				req = *tt.request
			}

			ctx := context.WithValue(
				context.Background(),
				loginTestContextKey{},
				"trace-uji",
			)

			repoCalls := 0
			jwtCalls := 0

			// Menggunakan fakeUserRepository dari test register.
			repo := &fakeUserRepository{
				checkEmailFn: func(
					gotCtx context.Context,
					email string,
				) (entities.User, bool, error) {
					repoCalls++

					if gotCtx != ctx {
						t.Error(
							"context tidak diteruskan ke repository",
						)
					}

					if email != "orwin@example.com" {
						t.Errorf(
							"email belum dinormalisasi: %q",
							email,
						)
					}

					if tt.repositoryErr != nil {
						return entities.User{}, false, tt.repositoryErr
					}

					if tt.missingUser {
						return entities.User{}, false, nil
					}

					return user, true, nil
				},
			}

			jwtStub := &loginJWTStub{
				generate: func(
					id string,
					role string,
				) string {
					jwtCalls++

					if id != user.ID.String() {
						t.Error(
							"user ID untuk JWT harus berasal dari database",
						)
					}

					if role != string(user.Role) {
						t.Error(
							"role untuk JWT harus berasal dari database",
						)
					}

					return "token-uji-bukan-jwt-asli"
				},
			}

			// Berbeda dari helper register:
			// login membutuhkan JWT service yang tidak nil.
			svc := service.NewAuthService(
				repo,
				jwtStub,
				nil,
				&stubRefreshRepo{},
			)

			result, gotErr := svc.Login(ctx, req)

			if repoCalls != tt.wantRepoCalls {
				t.Errorf(
					"pemanggilan repository: ingin %d, mendapat %d",
					tt.wantRepoCalls,
					repoCalls,
				)
			}

			if jwtCalls != tt.wantJWTCalls {
				t.Errorf(
					"pemanggilan JWT: ingin %d, mendapat %d",
					tt.wantJWTCalls,
					jwtCalls,
				)
			}

			// Periksa kasus gagal.
			if tt.wantErr != nil {
				if !errors.Is(gotErr, tt.wantErr) {
					t.Fatalf(
						"error: ingin %v, mendapat %v",
						tt.wantErr,
						gotErr,
					)
				}

				if result != (authDto.TokenResponse{}) {
					t.Error(
						"login gagal harus memberikan response kosong",
					)
				}

				return
			}

			// Periksa kasus berhasil.
			if gotErr != nil {
				t.Fatalf(
					"login seharusnya berhasil: %v",
					gotErr,
				)
			}

			want := authDto.TokenResponse{
				AccessToken: "token-uji-bukan-jwt-asli",
				Role:        string(user.Role),
			}

			if result.AccessToken != want.AccessToken || result.Role != want.Role {
				t.Errorf(
					"hasil login tidak sesuai:\nmendapat %+v\ningin %+v",
					result,
					want,
				)
			}
			if result.RefreshToken == "" {
				t.Error("refresh token harus terisi")
			}
		})
	}
}
