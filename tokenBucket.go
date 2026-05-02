package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	Capacity      float64
	RefillRate float64
	Current    float64
	lastUpdated time.Time
}

var tokenBucketrequests = make(map[string]TokenBucket)

func tokenBucketLimiter(ctx *gin.Context) {
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
	 
	 data,exists := tokenBucketrequests[user.UserID]
   
	 //initialize a full token bucket
	 if !exists{
		 tokenBucket := TokenBucket{
				Capacity:8,
				RefillRate:4,
				Current:8,
				lastUpdated: time.Now(),
			}
		 tokenBucket.Current-- //consumed on this request
		 tokenBucketrequests[user.UserID] = tokenBucket
		 mux.Unlock()
		 ctx.Next()
		 return
	 }

   data.updateTokenBucket() //now has updated tokens
	 fmt.Println(data.Current)
	 //token bucket present with enough tokens 
	 if data.Current >= 1.0{
		  data.Current = data.Current-1;
      tokenBucketrequests[user.UserID] = data //updated data is saved 
			mux.Unlock()
			ctx.Next()
			return
	 }

	 //already reached max value, need atleast 1 token to perform request
	 if data.Current < 1{
		mux.Unlock()
		ctx.AbortWithStatusJSON(429, gin.H{"message":"blocked, too many calls in the last minute"})
		return
	 }
}

func (tb *TokenBucket)updateTokenBucket(){
	seconds := time.Since(tb.lastUpdated).Seconds()
	tokensGained := (seconds)*(tb.RefillRate/60)
	fmt.Println(tokensGained,"tokens gained")
	if tokensGained + (tb.Current) > (tb.Capacity){
    tb.Current = tb.Capacity
		tb.lastUpdated = time.Now()
		return
	}
	tb.Current = tokensGained + tb.Current
	tb.lastUpdated = time.Now()
}
