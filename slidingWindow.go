package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type slidingBucket struct {
	count int
	slot  int64
}

var slidingWindowRequests = make(map[string][]slidingBucket) //all the requests for the users out there

func getBucketSlot() int64 {
	slot := time.Now().Unix() / 20
	return slot
}

// three windows inside a min
func slidingWindowLimiter(ctx *gin.Context) {
	user := new(User)
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"message": err})
		return
	}

	fmt.Println(user)

	if user.UserID == "" {
		ctx.AbortWithStatusJSON(400, gin.H{"message": "user id not found!"})
		return
	}

	mux.Lock()
	data, exists := slidingWindowRequests[user.UserID]
	bucketSlot := getBucketSlot()
	//initialize a full token bucket
	if !exists {
		bucketIndex := bucketSlot % 3
		data = make([]slidingBucket, 3)
		data[bucketIndex] = slidingBucket{
			count: 1,
			slot:  bucketSlot,
		}
		slidingWindowRequests[user.UserID] = data
		mux.Unlock()
		ctx.Next()
		return
	}

	activeRequests := getActiveRequestsInBuckets(data, bucketSlot)

	if activeRequests >= 8 {
		mux.Unlock()
		ctx.AbortWithStatusJSON(429, gin.H{"message": "too many requests in the last window"})
		return
	}

	if activeRequests < 8 {
		//update count for the correct slot
		getUpdatedBuckets(data, bucketSlot)
		slidingWindowRequests[user.UserID] = data
		mux.Unlock()
		ctx.Next()
	}
}

func getActiveRequestsInBuckets(buckets []slidingBucket, bucketSlot int64) int {
	count := 0
	for _, bucket := range buckets {
		if bucket.isActive(bucketSlot) {
			count += bucket.count
		}
	}
	return count
}

func (b slidingBucket) isActive(bucketSlot int64) bool {
	if bucketSlot-b.slot < 3 && bucketSlot-b.slot >= 0 {
		return true
	}
	return false
}

// slices are reference types so they update
func getUpdatedBuckets(buckets []slidingBucket, bucketSlot int64) {
	//dont do anything to inactive buckets
	bucketIndex := bucketSlot % 3

	//no such bucket exists with this slot so we create a new one
	if buckets[bucketIndex].slot != bucketSlot {
		buckets[bucketIndex] = slidingBucket{
			count: 0,
			slot:  bucketSlot,
		}
	}
	if buckets[bucketIndex].slot == bucketSlot {
		buckets[bucketIndex].count++
	}
}
