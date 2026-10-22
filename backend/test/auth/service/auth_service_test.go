package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type fakeUserRepository struct {
	checkEmailFn func(ctx context.Context, email string) (entities.User, bool, error)
	registerFn   func(ctx context.Context, user entities.User) (entities.User, error)

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
