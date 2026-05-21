package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ugcHandler(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			log.Println(err, string(debug.Stack()))
		}
	}()

	path := c.Param("path")

	if path == "/manifest.json" {
		ugcManifestHandler(c)
	} else {
		ugcFileHandler(c)
	}
}

func ugcManifestHandler(c *gin.Context) {
	id := c.Param("id")
	log.Println("ugcManifestHandler", c.Request.URL, id)

	path := getFilePath(id) + "manifest.json"

	_, err := os.Stat(path)

	notExist := errors.Is(err, os.ErrNotExist)

	log.Println("notExist", notExist)

	if err != nil && !notExist {
		log.Println("ugcManifestHandler", err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	if notExist {
		err = downloadFileIfNotExist(id)
		if err != nil {
			log.Println("failed to download file for item "+id, err)
			c.String(http.StatusInternalServerError, "")
		}

		generateItemManifest(id)
		if err != nil {
			log.Println("failed to generate manifest for item "+id, err)
			c.String(http.StatusInternalServerError, "")
		}
	}

	content, err := os.ReadFile(getFilePath(id) + "manifest.json")
	if err != nil {
		log.Println("failed to read manifest for item "+id, err)
		c.String(http.StatusInternalServerError, "")
	}

	c.Data(http.StatusOK, "application/json", content)
}

func ugcFileHandler(c *gin.Context) {
}

func getFilePath(id string) string {
	re := regexp.MustCompile(`(.{2}|.{1})`)
	x := re.FindAllStringSubmatch(id, -1)

	path := "./files"
	for _, s := range x {
		path += "/" + s[0]
	}

	path += "/" + id + "/"
	return path
}

func downloadFileIfNotExist(id string) error {
	_, err := os.Stat(getFilePath(id) + id + ".zip")

	notExist := errors.Is(err, os.ErrNotExist)
	if notExist {
		return downloadFile(id)
	}

	return nil
}

func downloadFile(id string) error {
	i, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}

	url, err := getItemUrl(i)
	if err != nil {
		return fmt.Errorf("database error when getting url for item "+id+": <%w>", err)
	}

	getPage(url)
	resp, err := getPage(url)
	if err != nil {
		return fmt.Errorf("failed to download file for item "+id+": <%w>", err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body for item "+id+": <%w>", err)
	}

	err = os.MkdirAll(getFilePath(id), 0755)
	if err != nil {
		return fmt.Errorf("failed to create folders for item "+id+" with path "+getFilePath(id)+": <%w>", err)
	}

	err = os.WriteFile(getFilePath(id)+id+".zip", buf, 0666)
	if err != nil {
		return fmt.Errorf("failed to write file for item "+id+": <%w>", err)
	}

	return nil
}

func generateItemManifest(id string) error {
	reader, err := zip.OpenReader(getFilePath(id) + id + ".zip")
	if err != nil {
		return fmt.Errorf("failed to open zip for item "+id+": <%w>", err)
	}
	defer reader.Close()

	files := []string{}
	for _, file := range reader.File {
		if true || strings.HasSuffix(file.Name, ".mdl") {
			files = append(files, file.Name)
		}
	}
	manifest := generateManifest(files)
	j, err := json.Marshal(&manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for item "+id+": <%w>", err)
	}
	err = os.WriteFile(getFilePath(id)+"manifest.json", j, 0666)
	if err != nil {
		return fmt.Errorf("failed to write manifest for item "+id+": <%w>", err)
	}

	return nil
}

func generateManifest(files []string) map[string]any {
	manifest := map[string]any{}
	for _, file := range files {
		path := strings.Split(file, "/")
		// Zip are supposed to only contain forward slashes, but this is not always the case
		if len(path) == 1 {
			path = strings.Split(file, "\\")
		}

		current := manifest
		l := len(path)
		for i, segment := range path {
			c, found := current[segment]
			if found {
				current = c.(map[string]any)
			} else {
				if i == l-1 {
					current[segment] = nil
				} else {
					c := map[string]any{}
					current[segment] = c
					current = c
				}
			}
		}
	}

	return manifest
}
