package providers

import (
	"github.com/samber/do"
	"github.com/story2gherkin/database"
	"github.com/story2gherkin/config"
	authHandler "github.com/story2gherkin/modules/auth/handler"
	authService "github.com/story2gherkin/modules/auth/service"
	userRepo "github.com/story2gherkin/modules/user/repository"
	"github.com/story2gherkin/pkg/constants"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(
		injector,
		constants.DB,
		func(i *do.Injector) (*gorm.DB, error) {
			db := config.SetUpDatabaseConnection()

			if err := database.Migrate(db); err != nil {
				return nil, err
			}

			return db, nil
		},
	)
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(
	injector,
	constants.JWTService,
	func(i *do.Injector) (authService.JWTService, error) {
			return authService.NewJWTService()
		},
	)

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := userRepo.NewUserRepository(db)

	authService := authService.NewAuthService(userRepository, jwtService, db)

	do.Provide(
		injector, func(i *do.Injector) (authHandler.AuthHandler, error) {
			return authHandler.NewAuthController(i, authService), nil
		},
	)
}