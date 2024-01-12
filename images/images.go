package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func getDatabaseConnection() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   "root",
		Passwd: "pass",
		Net:    "tcp",
		Addr:   "image_db:3306",
		DBName: "images",
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	return db, nil
}

// A UserFile represents a file uploaded by a user
type UserFile struct {
	UserID   int64  `json:"user_id"`
	Filename string `json:"filename"`
}

func addNewImageHandler(c *gin.Context) {
	userFile := &UserFile{}
	err := c.ShouldBind(&userFile)
	if err != nil {
		log.Printf("error binding request data: '%s'\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Printf("userFile data: %+v\n", userFile)

	queryString := "INSERT INTO images (user_id, filename) VALUES (?, ?)"
	res, err := db.Exec(queryString, userFile.UserID, userFile.Filename)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	log.Printf("inserted image '%s' with image id %d", userFile.Filename, id)
	c.JSON(http.StatusOK, gin.H{
		"image_id": id,
	})
}

type File struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
}

func getRecentImages(count int) ([]File, error) {
	queryString := "SELECT id, filename FROM images ORDER BY upload_date LIMIT ?"
	rows, err := db.Query(queryString, count)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var files []File

	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.Filename); err != nil {
			log.Println(err)
			return files, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); rows != nil {
		return files, err
	}
	return files, nil
}

func getRecentImagesHandler(c *gin.Context) {
	count := c.Param("count")
	numCount, err := strconv.Atoi(count)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if numCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of requested images must be positive."})
		return
	}

	recentImages, err := getRecentImages(numCount)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("recent: ", recentImages)
	c.JSON(http.StatusOK, gin.H{"images": recentImages})
}

type UserID struct{}

func getImagesUploadedByUser(userID int64) ([]string, error) {
	queryString := "SELECT filename FROM images WHERE user_id=? ORDER BY upload_date LIMIT 10"
	rows, err := db.Query(queryString, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var filenames []string

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			log.Println(err)
			return filenames, err
		}
		filenames = append(filenames, filename)
	}

	if err = rows.Err(); rows != nil {
		return filenames, err
	}
	return filenames, nil
}

func getImagesUploadedByUserHandler(c *gin.Context) {
	userID := c.Param("userID")
	numUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	images, err := getImagesUploadedByUser(numUserID)
	if err != nil {
		log.Println(err)
		log.Println(images)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"images": images})
}

func getImageByID(c *gin.Context) {
	imageIDString := c.Param("imageID")
	numImageID, err := strconv.ParseInt(imageIDString, 10, 64)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var filename string
	queryString := "SELECT filename FROM images WHERE id=? LIMIT 1"
	row := db.QueryRow(queryString, numImageID)
	if err := row.Scan(&filename); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found."})
			return
		}
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"filename": filename})
}

func main() {
	var err error
	db, err = getDatabaseConnection()
	if err != nil {
		log.Fatal(err)
		return
	} else {
		log.Println("Successfully connected to image database.")
	}

	r := gin.Default()

	r.POST("/add", addNewImageHandler)
	r.GET("/recent/:count", getRecentImagesHandler)
	r.GET("/user/:userID", getImagesUploadedByUserHandler)
	r.GET("/get/:imageID", getImageByID)

	r.Run(":8080")
}
