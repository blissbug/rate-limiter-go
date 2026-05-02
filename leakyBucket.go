package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var leakyBucketRequests = make(map[string]*leakyBucket)

var mu = new(sync.Mutex)

type leakyBucket struct {
	capacity     float64
	currentLevel float64
	stopTraffic  chan bool
	leakRate     float64
	updatedAt    time.Time
	mu           sync.Mutex
}

func newLeakyBucket() *leakyBucket {
	lb := &leakyBucket{}
	lb.capacity = 8
	lb.currentLevel = 0
	lb.leakRate = .1
	lb.stopTraffic = make(chan bool)
	lb.updatedAt = time.Now()
	lb.mu = sync.Mutex{}

	return lb
}

func (lb *leakyBucket) updateLeakyBucket() {
	passedTime := time.Since(lb.updatedAt).Seconds()
	leak := passedTime * lb.leakRate
	if lb.currentLevel-leak < 0 {
		lb.currentLevel = 0
	} else {
		lb.currentLevel = lb.currentLevel - leak
	}
	lb.updatedAt = time.Now()
	fmt.Println("Cleared and passed off!")
}

func leakyBucketLimiter(ctx *gin.Context) {
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
	mu.Lock()
	data, ok := leakyBucketRequests[user.UserID]

	if !ok {
		newLeakyBucket := newLeakyBucket()
		newLeakyBucket.mu.Lock()
		newLeakyBucket.currentLevel = 1
		newLeakyBucket.updatedAt = time.Now()
		newLeakyBucket.mu.Unlock()
		leakyBucketRequests[user.UserID] = newLeakyBucket
		mu.Unlock()
		ctx.Next()
		return
	}

	mu.Unlock()

	data.mu.Lock()
	data.updateLeakyBucket()
	if data.capacity-data.currentLevel < 1 {
		data.mu.Unlock()
		ctx.AbortWithStatusJSON(429, gin.H{"message": "too many requests in the last window"})
		return
	} else {
		//update count for the correct slot
		data.currentLevel++
		data.mu.Unlock()
		ctx.Next()
		return
	}
}
