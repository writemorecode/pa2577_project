package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Image Sharing Service",
	})
}

func imageUploadHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.String(http.StatusBadRequest, "get form error: %s", err.Error())
		return
	}

	filename := filepath.Base(file.Filename)
	fileExtension := filepath.Ext(filename)
	uuidFilename := uuid.New().String() + fileExtension
	uploadPath := filepath.Join("/images", uuidFilename)

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.String(http.StatusBadRequest, "upload form error: %s", err.Error())
		return
	}
	c.Redirect(http.StatusFound, uploadPath)
}

func main() {
	r := gin.Default()

	r.Static("/images", "./images")

	r.LoadHTMLGlob("*.tmpl")

	// Max uploaded image size 8 MiB
	r.MaxMultipartMemory = 8 << 20

	r.GET("/", indexHandler)
	r.POST("/upload", imageUploadHandler)
	r.Run(":8080")
}
