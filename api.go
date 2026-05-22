package main

import (
	"encoding/gob"
	"errors"
	"log"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

var _ = registerToken()

const ugcUrlPrefix = "https://images.steamusercontent.com/ugc/"

func registerToken() bool {
	gob.Register(map[string]any{})
	gob.Register(map[string]bool{})
	gob.Register(struct{}{})
	return true
}

type apiRequest struct {
	Action  string         `json:"action" binding:"required"`
	Version int            `json:"version" binding:"required"`
	Params  map[string]any `json:"params"`
}

func apiHandler(c *gin.Context) {
	var request apiRequest
	var err error

	defer func() {
		if err := recover(); err != nil {
			jsonError(c, CreateApiError(UnexpectedError))
			log.Println(err, string(debug.Stack()))
		}
	}()

	if err = c.ShouldBindJSON(&request); err != nil {
		logError(c, err)
		jsonError(c, errors.New("bad request"))
		return
	}

	var apiError apiError
	switch request.Action {
	case "get-item-list":
		apiError = apiGetItemList(c, request.Params)
	default:
		jsonError(c, NotFoundError{})
		return
	}

	if apiError != nil {
		jsonError(c, apiError)
	}
}

func apiGetItemList(c *gin.Context, params map[string]any) apiError {
	result, err := getItems()

	if err != nil {
		logError(c, err)
		return CreateApiError(UnexpectedError)
	}

	for i := range result {
		result[i].PreviewUrl = strings.Replace(result[i].PreviewUrl, ugcUrlPrefix, "/", 1)
	}

	jsonSuccess(c, map[string]any{"items": result})
	return nil
}
