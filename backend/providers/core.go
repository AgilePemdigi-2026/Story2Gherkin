package providers

import (
	"github.com/samber/do"
	"github.com/story2gherkin/config"
	authHandler "github.com/story2gherkin/modules/auth/handler"
	authService "github.com/story2gherkin/modules/auth/service"
	userRepo "github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/constants"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := userRepo.NewUserRepository(db)

	mailer := authService.NewSMTPPasswordResetMailer()

	authService := authService.NewAuthService(
		userRepository,
		jwtService,
		db,
		mailer,
	)

	do.Provide(
		injector, func(i *do.Injector) (authHandler.AuthHandler, error) {
			return authHandler.NewAuthController(i, authService), nil
		},
	)
}
