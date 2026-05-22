package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var LockExists = errors.New("lock exists")

const downloadRetries = 10
const downloadRetryAfter = 10 * time.Second

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
	defer func() {
		if err := recover(); err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			log.Println(err, string(debug.Stack()))
		}
	}()

	id := c.Param("id")

	path := getManifestPath(id)

	_, err := os.Stat(path)

	notExist := errors.Is(err, os.ErrNotExist)

	if err != nil && !notExist {
		log.Println("error in ugcManifestHandler accessing path "+path, err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	if notExist {
		err = downloadFileIfNotExist(id)
		if err != nil {
			log.Println("failed to download file for item "+id, err)
			c.String(http.StatusInternalServerError, "")
			return
		}

		err = generateItemManifest(id)
		if err != nil {
			log.Println("failed to generate manifest for item "+id, err)
			c.String(http.StatusInternalServerError, "")
			return
		}
	}

	content, err := os.ReadFile(getManifestPath(id))
	if err != nil {
		log.Println("failed to read manifest for item "+id, err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}

func ugcFileHandler(c *gin.Context) {
	id := c.Param("id")
	path := c.Param("path")
	//path = "models/mysterypancake/../../....\\..\\/../../..\\androidlogo/android.mdl"
	//path = "/../getprintfulproducts.json"

	// Cleanup the path
	path = filepath.Clean("." + path)

	// Check if the path is local
	if !filepath.IsLocal(path) {
		log.Println("path is not local " + path + " for item " + id)
		c.String(http.StatusInternalServerError, "")
		return
	}

	log.Println(path)

	_, err := os.Stat(path)

	notExist := errors.Is(err, os.ErrNotExist)

	if err != nil && !notExist {
		log.Println("error in ugcFileHandler accessing path "+path, err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	if notExist {
		err = extractFile(id, path)
		if err != nil {
			log.Println("error in ugcFileHandler "+path, err)
			c.String(http.StatusInternalServerError, "")
			return
		}
	}

	completePath := filepath.Join(getFilesPath(id), path)

	content, err := os.ReadFile(completePath)
	if err != nil {
		log.Println("failed to read path +"+path+" for item "+id, err)
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)

	//c.Data(http.StatusOK, "application/json", content)
}

func getItemPath(id string) string {
	re := regexp.MustCompile(`(.{2}|.{1})`)
	x := re.FindAllStringSubmatch(id, -1)

	path := "./files"
	for _, s := range x {
		path += "/" + s[0]
	}

	path += "/" + id + "/"
	return path
}

func getManifestPath(id string) string {
	return getItemPath(id) + "manifest.json"
}

func getZipPath(id string) string {
	return getItemPath(id) + id + ".zip"
}

func getLockPath(id string) string {
	return getItemPath(id) + "download.lock"
}

func getFilesPath(id string) string {
	return path.Join(getItemPath(id), "files")
}

func downloadFileIfNotExist(id string) error {
	// This function check if the file exists before trying to download it
	dl := func() error {
		_, err := os.Stat(getZipPath(id))

		notExist := errors.Is(err, os.ErrNotExist)
		if notExist {
			return downloadFile(id)
		}
		return nil
	}

	for range downloadRetries {
		err := dl()
		if err != LockExists {
			return err
		}

		//
		time.Sleep(downloadRetryAfter)
	}

	return errors.New("failed to download file " + id + " after " + strconv.Itoa(10) + " retries")
}

func downloadFile(id string) error {
	// Check if the download lock exists
	_, err := os.Stat(getLockPath(id))
	if err == nil {
		// lock exists, return
		return LockExists
	}

	// Create the directory
	err = os.MkdirAll(getFilesPath(id), 0755)
	if err != nil {
		return fmt.Errorf("failed to create folders for item "+id+" with path "+getFilesPath(id)+": <%w>", err)
	}

	// Create the lock
	_, err = os.Create(getLockPath(id))
	if err != nil {
		return errors.New("failed to create a lock file " + getLockPath(id) + " for item " + id)
	}
	defer os.Remove(getLockPath(id))

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

	err = os.WriteFile(getZipPath(id), buf, 0666)
	if err != nil {
		return fmt.Errorf("failed to write file for item "+id+": <%w>", err)
	}

	return nil
}

func extractFile(id string, p string) error {
	err := downloadFileIfNotExist(id)
	if err != nil {
		return fmt.Errorf("failed to extract file for item "+id+": <%w>", err)
	}

	reader, err := zip.OpenReader(getZipPath(id))
	if err != nil {
		return fmt.Errorf("failed to open zip for item "+id+": <%w>", err)
	}
	defer reader.Close()

	file, err := reader.Open(p)
	if err != nil {
		return fmt.Errorf("failed to open zip file "+p+" for item "+id+": <%w>", err)
	}

	defer file.Close()

	buf, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return fmt.Errorf("failed to read zip file "+p+" for item "+id+": <%w>", err)
	}

	completePath := filepath.Join(getFilesPath(id), p)

	err = os.MkdirAll(path.Dir(completePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create folders for item "+id+" with path "+completePath+": <%w>", err)
	}

	err = os.WriteFile(completePath, buf, 0666)
	if err != nil {
		return fmt.Errorf("failed to write file for item "+id+": <%w>", err)
	}

	return nil
}

func generateItemManifest(id string) error {
	reader, err := zip.OpenReader(getZipPath(id))
	if err != nil {
		return fmt.Errorf("failed to open zip for item "+id+": <%w>", err)
	}
	defer reader.Close()

	files := []string{}
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".mdl") || strings.HasSuffix(file.Name, ".pcf") {
			files = append(files, file.Name)
		}
	}
	manifest := generateManifest(files)
	j, err := json.Marshal(&manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for item "+id+": <%w>", err)
	}
	err = os.WriteFile(getManifestPath(id), j, 0666)
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
					current[segment] = 1
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
