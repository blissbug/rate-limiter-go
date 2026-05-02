package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

//in a minute we allow max of 8 requests
type timeSlice []time.Time

//implement sortable with the three functions
func (t timeSlice)Len()int{
	return len(t)
}

func (t timeSlice)Swap(i,j int){
   t[i],t[j] = t[j],t[i]
}

func (t timeSlice)Less(i,j int)bool{
   return t[i].Before(t[j])
}
var slidingLogRequests = make(map[string]timeSlice)

func slidingLogLimiter(ctx *gin.Context) {
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
	 data,exists := slidingLogRequests[user.UserID]
   
	 //initialize a full token bucket
	 if !exists{
		 slidingLogRequests[user.UserID] = append(data,time.Now())
		 mux.Unlock()
		 ctx.Next()
		 return
	 }

	 //add updater function which clears up before we perform more actions 
	 updatedTimeStamps := updateRequestStatus(data)

	 if len(updatedTimeStamps) >= 8{
			//reject the request
			slidingLogRequests[user.UserID] = append(updatedTimeStamps,time.Now())
			mux.Unlock()
			ctx.AbortWithStatusJSON(429,gin.H{"message":"too many requests in the last window"})
			return
	 }
   
	 //inside the rate limit
	 if len(updatedTimeStamps)<=8{
     slidingLogRequests[user.UserID] = append(updatedTimeStamps,time.Now())
		 mux.Unlock()
		 ctx.Next()
		 return
	 }
}

func updateRequestStatus(timestamps timeSlice)(timeSlice){
	var result timeSlice
	for _,timeStamp := range timestamps{
		if time.Since(timeStamp) < time.Minute{
       result = append(result,timeStamp)
		}
	}
	return result
}