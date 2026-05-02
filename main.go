package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//r.Use(FixedWindowRateLimiter) // fixed window rate limiter
	//r.Use(tokenBucketLimiter)
	//r.Use(slidingLogLimiter)
	//r.Use(slidingWindowLimiter)
	//r.Use(leakyBucketLimiter)
	r.GET("/home", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Request was suprisingly successful!"})
	})

	r.Run()
}
