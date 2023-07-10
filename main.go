package main

import (
        "bytes"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "strconv"
        "path/filepath"

        "github.com/gin-contrib/cors"
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

        apiURL := "http://localhost:8081/bot%s/%s"
        bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(token, apiURL)
        //bot, err := tgbotapi.NewBotAPI(token)
        if err != nil {
                log.Panic(err)
        }

        bot.Debug = true

        log.Printf("Authorized on account %s", bot.Self.UserName)

        r := gin.Default()

        r.Use(cors.New(cors.Config{
                AllowAllOrigins: true,
        }))

        r.GET("/ping", func(c *gin.Context) {
                c.JSON(200, gin.H{
                        "message": "pong",
                })
        })

        r.GET("/storage/:id", func(c *gin.Context) {
                fileID := c.Param("id")
                file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
                if err != nil {
                        log.Fatal(err)
                        c.AbortWithStatus(http.StatusNotFound)
                        return
                }

                localFilePath := file.FilePath

                fileData, err := ioutil.ReadFile(localFilePath)
                if err != nil {

                        c.AbortWithStatus(http.StatusInternalServerError)
                        return
                }

                contentLength := int64(len(fileData))
                contentType := http.DetectContentType(fileData)
                fileName := filepath.Base(localFilePath)

                extraHeaders := map[string]string{
                        "Content-Disposition": fmt.Sprintf("attachment; filename=%s", fileName),
                }

                c.DataFromReader(http.StatusOK, contentLength, contentType, bytes.NewReader(fileData), extraHeaders)
        })
        // Upload file to storage
        r.POST("/storage", func(c *gin.Context) {
                fileHeader, err := c.FormFile("file")

                if err != nil {
                        c.AbortWithStatus(http.StatusBadRequest)
                        return
                }

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
