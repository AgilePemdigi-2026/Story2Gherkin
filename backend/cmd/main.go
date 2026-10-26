package main

import (
	"log"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/story2gherkin/middlewares"
	"github.com/story2gherkin/modules/auth"
	"github.com/story2gherkin/modules/scenario"
	"github.com/story2gherkin/providers"
)

func run(server *gin.Engine) {
	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "0.0.0.0:" + port
	} else {
		serve = ":" + port
	}

	myFigure := figure.NewColorFigure("Story2Gherkin", "doom", "green", true)
	myFigure.Print()

	err := server.Run(serve)
	if err != nil {
		log.Fatal("Error running server: ", err)
	}
}

func main() {
	var (
		injector = do.New()
	)

	providers.RegisterDependencies(injector)

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	v1 := server.Group("api/v1")
	{
		scenario.RegisterRoutes(v1, injector)
		auth.RegisterRoutes(v1, injector)
	}

	run(server)
}
