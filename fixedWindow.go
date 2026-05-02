package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var mux sync.Mutex
var fixedWindowRequests = make(map[string]windowData)

type windowData struct{
   count int 
	 windowStart time.Time
}

type User struct{
	UserID string `json:"userID"`
}

//handles lock and unlock when reading/writing from dictionary 
func FixedWindowRateLimiter(ctx *gin.Context) {
	 user := new(User)
	 err := ctx.ShouldBindJSON(&user)
	 if err != nil{
		 ctx.AbortWithStatusJSON(400,gin.H{"message":err})
     return
	 }

	 fmt.Println(user)


	 if user.UserID == ""{
		 ctx.AbortWithStatusJSON(400,gin.H{"message":"user id not found!"})
     return
	 }

	 mux.Lock()
	 
	 data,exists := fixedWindowRequests[user.UserID]

	 if !exists{
		 fixedWindowRequests[user.UserID] = windowData{1,time.Now()}
		 mux.Unlock()
		 ctx.Next()
		 return
	 }

	 //time since last request was more than a minute
	 if time.Since(data.windowStart) > 1 * time.Minute{
      fixedWindowRequests[user.UserID] = windowData{1,time.Now()}
			mux.Unlock()
			ctx.Next()
			return
	 }
	 //already reached max value
	 if data.count >= 5{
		mux.Unlock()
		ctx.AbortWithStatusJSON(429, gin.H{"message":"blocked, too many calls in the last minute"})
		return
	 }
	 //update count
	 data.count++
	 fixedWindowRequests[user.UserID] = data
	 mux.Unlock()
	 ctx.Next()
}