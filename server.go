package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gomail "gopkg.in/gomail.v2"
)

type requestData struct {
	Name    string
	Email   string
	Message string
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := gin.Default()
	r.Use(gin.Logger())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"reply": "pong"})
	})
	r.POST("/send_mail", mailer)

	r.Run(":" + port)
}

func mailer(c *gin.Context) {
	log.Println(c.MultipartForm())
	// load env
	dotenvErr := godotenv.Load()
	if dotenvErr != nil {
		log.Fatal("Error loading .env file")
	}

	// setup template
	t := template.New("template.html")

	var err error
	t, err = t.ParseFiles("template.html")
	if err != nil {
		log.Println(err)
	}

	requestData := requestData{c.PostForm("name"), c.PostForm("email"), c.PostForm(("message"))}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, requestData); err != nil {
		log.Println(err)
	}
	result := tpl.String()

	// setup file
	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set Folder untuk menyimpan filenya
	path := file.Filename
	log.Println("path: ", path)
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// setup mailer
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("CONFIG_SENDER_NAME"))
	mailer.SetHeader("To", os.Getenv("CONFIG_RECIPIENT_NAME"))
	mailer.SetHeader("Subject", "MAILER GO")
	mailer.SetBody("text/html", result)
	mailer.Attach(path)

	port, atoiErr := strconv.Atoi(os.Getenv("CONFIG_SMTP_PORT"))
	if atoiErr != nil {
		port = 587
	}

	dialer := gomail.NewDialer(
		os.Getenv("CONFIG_SMTP_HOST"),
		port,
		os.Getenv("CONFIG_AUTH_EMAIL"),
		os.Getenv("CONFIG_AUTH_PASSWORD"),
	)

	dialErr := dialer.DialAndSend(mailer)
	if dialErr != nil {
		log.Fatal(dialErr.Error())
	}

	log.Println("Mail sent!")
	c.JSON(http.StatusOK, gin.H{"success": "ok"})
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
}
