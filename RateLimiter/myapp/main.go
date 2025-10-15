//Use gin for building backend

// filter the content of the request
package main

import (
	"context"
	"fmt"
	"net/http"
	"rate-limiter/mypackage"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func main() {

	app := gin.Default() //Create a app server
	//Make redis connection here
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	//Ping the redis to check it is working or  not
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error in connceting the redis", err)
	}
	fmt.Println("Redis connected ")

	AggressiveRateLimiter := mypackage.CheckCountValue(RedisClient, 10, 1)
	// End point for the incoming request
	app.GET("/resoruces", AggressiveRateLimiter, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully entered",
		})

	})

	//Just check the field of the request

	app.Run("127.0.0.1:3000")
}
