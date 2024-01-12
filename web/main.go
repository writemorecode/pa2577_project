package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageID struct {
	ImageID int64 `json:"image_id"`
}

func imageUploadHandler(c *gin.Context) {
	userIDString, err := c.Cookie("userID")
	if err == http.ErrNoCookie {
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}
	numUserID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusFound, "/")
		return
	}

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

	userFile := &UserFile{
		UserID:   numUserID,
		Filename: uuidFilename,
	}
	log.Printf("Sending %+v to images service", userFile)
	userFileJSON, err := json.Marshal(userFile)
	bodyReader := bytes.NewReader(userFileJSON)

	res, err := http.Post("http://images:8080/add", "application/json", bodyReader)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	if res.StatusCode != http.StatusOK {
		log.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Login failed."})
		return
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	var imageID ImageID
	err = json.Unmarshal(respBody, &imageID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	if imageID.ImageID <= 0 {
		log.Printf("Error: Image service returned invalid image ID: %d\n", imageID.ImageID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}
	log.Println("recieved image id from image service: ", imageID.ImageID)

	imagePath := fmt.Sprintf("/img/%d", imageID.ImageID)
	c.Redirect(http.StatusFound, imagePath)
}

type File struct {
	Filename string `json:"filename"`
}

func imageViewHandler(c *gin.Context) {
	imageID := c.Param("imageID")

	url := fmt.Sprintf("http://images:8080/get/%s", imageID)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}
	if res.StatusCode == http.StatusNotFound {
		c.HTML(http.StatusNotFound, "404", gin.H{})
		return
	}
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	var file File
	err = json.Unmarshal(respBody, &file)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	log.Printf("image service returned filename: '%s'\n", file.Filename)

	c.HTML(http.StatusOK, "view_image", gin.H{
		"filename": file.Filename,
	})
}

type User struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}
type UserSession struct {
	ID       int64
	Username string
}
type UserFile struct {
	UserID   int64  `json:"user_id"`
	Filename string `json:"filename"`
}

func loginHandler(c *gin.Context) {
	var user User
	err := c.ShouldBind(&user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		return
	}
	bodyBytes, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		return
	}
	bodyReader := bytes.NewReader(bodyBytes)

	endpointURL := url.URL{
		Host:   "auth:8080",
		Scheme: "http",
		Path:   "/login",
	}
	resp, err := http.Post(endpointURL.String(), "application/json", bodyReader)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Login failed."})
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	var userSession UserSession
	err = json.Unmarshal(respBody, &userSession)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	c.SetCookie("userID", fmt.Sprintf("%d", userSession.ID), 3600, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/")
}

func registerHandler(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		return
	}
	bodyBytes, err := json.Marshal(user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		return
	}
	bodyReader := bytes.NewReader(bodyBytes)

	endpointURL := url.URL{
		Host:   "auth:8080",
		Scheme: "http",
		Path:   "/register",
	}
	resp, err := http.Post(endpointURL.String(), "application/json", bodyReader)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad gateway."})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "User registration failed."})
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error."})
		return
	}

	var userSession UserSession
	err = json.Unmarshal(respBody, &userSession)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal server error."})
		return
	}

	c.SetCookie("userID", fmt.Sprintf("%d", userSession.ID), 3600, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/")
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("userID", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

func createTemplateRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "base.html", "index.html")
	r.AddFromFiles("upload", "base.html", "upload.html")
	r.AddFromFiles("view_image", "base.html", "view_image.html")
	r.AddFromFiles("404", "base.html", "404.html")
	r.AddFromFiles("500", "base.html", "500.html")
	r.AddFromFiles("login", "base.html", "login.html")
	r.AddFromFiles("register", "base.html", "register.html")
	return r
}

type Image struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
}
type ImageList struct {
	Images []Image `json:"images"`
}

func getRecentImages(count int) ([]Image, error) {
	url := fmt.Sprintf("http://images:8080/recent/%d", count)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var imageList ImageList

	err = json.Unmarshal(body, &imageList)
	if err != nil {
		return nil, err
	}

	return imageList.Images, nil
}

func getUserImages(userID string) ([]Image, error) {
	url := fmt.Sprintf("http://images:8080/user/%d", userID)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var imageList ImageList

	err = json.Unmarshal(body, &imageList)
	if err != nil {
		return nil, err
	}

	return imageList.Images, nil
}

func homePageHandler(c *gin.Context) {
	var err error
	userID, err := c.Cookie("userID")
	isLoggedIn := (err != http.ErrNoCookie)

	var images []Image
	imageCount := 4
	if isLoggedIn {
		images, err = getUserImages(userID)
	} else {
		images, err = getRecentImages(imageCount)
	}
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusInternalServerError, "500", gin.H{})
		return
	}
	c.HTML(http.StatusOK, "index", gin.H{
		"isLoggedIn": isLoggedIn,
		"images":     images,
	})
	return
}

func main() {
	r := gin.Default()
	r.HTMLRender = createTemplateRenderer()
	r.Static("/images", "./images")

	// Max uploaded image size 8 MiB
	r.MaxMultipartMemory = 8 << 20

	r.GET("/", homePageHandler)

	r.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload", gin.H{})
	})
	r.POST("/upload", imageUploadHandler)

	r.GET("/img/:imageID", imageViewHandler)
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login", gin.H{})
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register", gin.H{})
	})
	r.POST("/auth/login", loginHandler)
	r.GET("/logout", logoutHandler)
	r.POST("/auth/register", registerHandler)

	r.Run(":8080")
}
