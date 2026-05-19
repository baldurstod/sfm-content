package main

import (
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

func logError(c *gin.Context, e error) {
	_, file, line, ok := runtime.Caller(1)
	var requestID = ""
	if c != nil {
		requestID = c.GetHeader("X-Request-ID")
	}
	if ok {
		log.Println(requestID, file, line, e)
	} else {
		log.Println(requestID, e)
	}
}
