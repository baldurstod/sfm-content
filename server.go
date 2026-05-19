package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func startServer(config Config) {
	engine := initEngine(config)
	var err error

	log.Printf("Listening on port %d\n", config.Port)
	err = engine.RunTLS(":"+strconv.Itoa(config.Port), config.HttpsCertFile, config.HttpsKeyFile)
	log.Fatal(err)
}

func initEngine(config Config) *gin.Engine {
	if ReleaseMode == "true" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowMethods:    []string{"POST", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Request-Id"},
		AllowAllOrigins: true,
		MaxAge:          12 * time.Hour,
	}))

	return r
}
