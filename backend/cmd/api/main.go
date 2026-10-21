package main

import (
	"log"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
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
	server := gin.Default()

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	run(server)
}
