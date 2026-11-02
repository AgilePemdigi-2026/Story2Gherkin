package bdd_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type bddUserRepo struct {
	emails       map[string]bool
	checkErr     error
	registerErr  error
	checkedEmail string
	registered   *entities.User
	registerCall int
}

func (f *bddUserRepo) CheckEmail(_ context.Context, _ *gorm.DB, email string) (entities.User, bool, error) {
	f.checkedEmail = email
	if f.checkErr != nil {
		return entities.User{}, false, f.checkErr
	}
	return entities.User{}, f.emails[email], nil
}

func (f *bddUserRepo) Register(_ context.Context, _ *gorm.DB, user entities.User) (entities.User, error) {
	f.registerCall++
	f.registered = &user
	if f.registerErr != nil {
		return entities.User{}, f.registerErr
	}
	return user, nil
}

func (f *bddUserRepo) GetUserByEmail(context.Context, *gorm.DB, string) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (f *bddUserRepo) GetUserById(context.Context, *gorm.DB, string) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (f *bddUserRepo) Update(_ context.Context, _ *gorm.DB, user entities.User) (entities.User, error) {
	return user, nil
}

func (f *bddUserRepo) Delete(context.Context, *gorm.DB, string) error { return nil }

type registerServiceFeature struct {
	repo   *bddUserRepo
	result userDto.UserResponse
	err    error
}

func (f *registerServiceFeature) tidakAdaPengguna(email string) error {
	delete(f.repo.emails, email)
	f.repo.checkErr = nil
	f.repo.registerErr = nil
	return nil
}

func (f *registerServiceFeature) sudahAdaPengguna(email string) error {
	f.repo.emails[email] = true
	return nil
}

func (f *registerServiceFeature) cekEmailError(msg string) error {
	f.repo.checkErr = fmt.Errorf("%s", msg)
	return nil
}

func (f *registerServiceFeature) tidakAdaDanInsertError(email, msg string) error {
	delete(f.repo.emails, email)
	f.repo.checkErr = nil
	f.repo.registerErr = fmt.Errorf("%s", msg)
	return nil
}

func (f *registerServiceFeature) registerService(name, email, password, role string) error {
	svc := service.NewAuthService(f.repo, nil, nil, nil)
	f.result, f.err = svc.Register(context.Background(), userDto.UserCreateRequest{
		Name: name, Email: email, Password: password, Role: role,
	})
	return nil
}

func (f *registerServiceFeature) registerPanjang(name, email string, n int, role string) error {
	return f.registerService(name, email, strings.Repeat("a", n), role)
}

func (f *registerServiceFeature) berhasilDengan(name, email, role string) error {
	if f.err != nil {
		return fmt.Errorf("register error = %v, ingin sukses", f.err)
	}
	if f.result.Name != name || f.result.Email != email || f.result.Role != role || f.result.ID == "" {
		return fmt.Errorf("respons = %+v, ingin name=%q email=%q role=%q id non-kosong", f.result, name, email, role)
	}
	if f.repo.registered == nil {
		return fmt.Errorf("repository.Register tidak dipanggil")
	}
	if f.repo.registered.Name != name || f.repo.registered.Email != email || string(f.repo.registered.Role) != role {
		return fmt.Errorf("tersimpan = %+v, ingin name=%q email=%q role=%q", f.repo.registered, name, email, role)
	}
	if f.repo.registered.Password == "" || f.repo.registered.Password == "rahasia123" {
		return fmt.Errorf("password tersimpan plaintext/kosong")
	}
	plain := "rahasia123"
	if email == "rizal@example.com" && name == "Rizal" {
		if err := bcrypt.CompareHashAndPassword([]byte(f.repo.registered.Password), []byte(plain)); err != nil {
			return fmt.Errorf("password hash tidak cocok: %v", err)
		}
	}
	return nil
}

func (f *registerServiceFeature) gagalDuplikat() error {
	if f.err == nil || f.err.Error() != userDto.ErrEmailAlreadyExists.Error() {
		return fmt.Errorf("error = %v, ingin %v", f.err, userDto.ErrEmailAlreadyExists)
	}
	if f.repo.registerCall != 0 {
		return fmt.Errorf("Register dipanggil %d kali, ingin 0", f.repo.registerCall)
	}
	return nil
}

func (f *registerServiceFeature) gagalDengan(msg string) error {
	if f.err == nil || f.err.Error() != msg {
		return fmt.Errorf("error = %v, ingin %q", f.err, msg)
	}
	return nil
}

func (f *registerServiceFeature) gagalTanpaRegister() error {
	if f.err == nil {
		return fmt.Errorf("error = nil, ingin gagal")
	}
	if f.repo.registerCall != 0 {
		return fmt.Errorf("Register dipanggil %d kali, ingin 0", f.repo.registerCall)
	}
	return nil
}

func InitializeRegisterServiceScenario(ctx *godog.ScenarioContext) {
	f := &registerServiceFeature{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		f.repo = &bddUserRepo{emails: map[string]bool{}}
		f.result = userDto.UserResponse{}
		f.err = nil
		return ctx, nil
	})
	ctx.Step(`^tidak ada pengguna beremail "([^"]*)"$`, f.tidakAdaPengguna)
	ctx.Step(`^sudah ada pengguna beremail "([^"]*)"$`, f.sudahAdaPengguna)
	ctx.Step(`^cek email error "([^"]*)"$`, f.cekEmailError)
	ctx.Step(`^tidak ada pengguna beremail "([^"]*)" dan insert error "([^"]*)"$`, f.tidakAdaDanInsertError)
	ctx.Step(`^saya register service dengan nama "([^"]*)", email "([^"]*)", password "([^"]*)", role "([^"]*)"$`, f.registerService)
	ctx.Step(`^saya register service dengan nama "([^"]*)", email "([^"]*)", password sepanjang (\d+) karakter, role "([^"]*)"$`, f.registerPanjang)
	ctx.Step(`^register service berhasil dengan nama "([^"]*)", email "([^"]*)", role "([^"]*)" dan password ter-hash$`, f.berhasilDengan)
	ctx.Step(`^register service gagal email sudah ada dan repository Register tidak dipanggil$`, f.gagalDuplikat)
	ctx.Step(`^register service gagal dengan error "([^"]*)"$`, f.gagalDengan)
	ctx.Step(`^register service gagal dan repository Register tidak dipanggil$`, f.gagalTanpaRegister)
}

func TestRegisterServiceFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeRegisterServiceScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/auth_register_service.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
