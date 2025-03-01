package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func endpoint(c *gin.Context) {
	addr := c.Request.RemoteAddr
	log.Println("receive request from: " + addr)
}

func middlewareLogger(f func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		log.Println("start enhance endpoint")
		f(c)
		log.Println("end enhance endpoint")
	}
}

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.Request.Header.Get("access_token")
		if token != "test" {
			context.JSON(http.StatusForbidden, gin.H{
				"message": "auth failed with wrong token",
			})
			context.Abort()
		}
		log.Println("auth success")
	}
}

func main() {
	r := gin.Default()
	r.Use(gin.Logger(), gin.Recovery(), middlewareLogger(endpoint), Auth())
	dsware := r.Group("/dsware")
	dsware.GET("remake", func(context *gin.Context) {
		context.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "don't give up",
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
