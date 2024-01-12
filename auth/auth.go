package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}

var (
	db *sql.DB
)

func getDatabaseConnection() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   "root",
		Passwd: "pass",
		Net:    "tcp",
		Addr:   "user_db:3306",
		DBName: "users",
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Println(pingErr)
		return nil, err
	}

	return db, nil
}

func lookupUser(c *gin.Context) {
	id := c.Param("id")
	var username string
	row := db.QueryRow("SELECT username FROM users WHERE id=?", id)
	err := row.Scan(&username)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"username": username})
}

func registerUser(c *gin.Context) {
	user := &User{}
	err := c.ShouldBindJSON(&user)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := db.Ping(); err != nil {
		log.Println("ERROR: Could not ping database.")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("USER: %+v\n", user)

	var id int64
	row := db.QueryRow("SELECT id FROM users WHERE username=?", user.Username)

	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		break
	case nil:
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	default:
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.PasswordHash)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	id, _ = result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"id":       id,
		"username": user.Username,
	})

}

func loginUser(c *gin.Context) {
	user := &User{}
	err := c.ShouldBind(&user)
	if err != nil {
		log.Printf("error binding request data: '%s'\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Println(user)

	row := db.QueryRow("SELECT id FROM users WHERE username=? AND password=?", user.Username, user.PasswordHash)

	var userID int64
	err = row.Scan(&userID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"id": userID, "username": user.Username})
		return
	} else if err == sql.ErrNoRows {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not found"})
		return
	} else {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func main() {
	var err error
	db, err = getDatabaseConnection()
	if err != nil {
		log.Println("ERROR: FAILED TO CONNECT TO USER DATABASE!")
		log.Fatal(err)
		return
	} else {
		log.Println("Successfully connected to user database.")
	}

	r := gin.Default()
	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/lookup/:id", lookupUser)
	r.Run(":8080")
}
