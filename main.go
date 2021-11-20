package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	token := os.Getenv("TELEGRAM_TOKEN")
	chatId := os.Getenv("TELEGRAM_CHAT_ID")

	// Check if token is set else throw error with message
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}

	// Check if chatId is set else throw error with message
	if chatId == "" {
		log.Fatal("TELEGRAM_CHAT_ID is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/storage/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")

		directUrl, err := bot.GetFileDirectURL(id)

		fmt.Print(directUrl)

		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// download and serve directUrl
		resp, err := http.Get(directUrl)

		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		contentLength := resp.ContentLength
		contentType := resp.Header.Get("Content-Type")

		// get filename from directUrl
		fileName := directUrl[strings.LastIndex(directUrl, "/")+1:]

		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf("attachment; filename=%s", fileName),
		}

		defer resp.Body.Close()

		c.DataFromReader(http.StatusOK, contentLength, contentType, resp.Body, extraHeaders)
	})

	// Upload file to storage
	r.POST("/storage", func(c *gin.Context) {
		fileHeader, _ := c.FormFile("file")
		file, _ := fileHeader.Open()
		bytes, _ := ioutil.ReadAll(file)

		upload_bytes := tgbotapi.FileBytes{Name: fileHeader.Filename, Bytes: bytes}

		chatId, _ := strconv.ParseInt(chatId, 10, 64)

		msg := tgbotapi.NewDocument(chatId, upload_bytes)
		ret, err := bot.Send(msg)

		if err != nil {
			log.Panic(err)
		}

		c.JSON(200, ret.Document)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
