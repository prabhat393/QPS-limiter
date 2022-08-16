package main

import (
	"fmt"
	"log"
	"net/http"
	"okapi-public-api/pkg/localmw"
	"strconv"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	gin.SetMode(os.Getenv("API_MODE"))

	_ = os.Setenv("TZ", "UTC")

	router := gin.New()

	cfg := &redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	}

	cmd := redis.NewClient(cfg)

	qpsLimit, _ := strconv.Atoi(os.Getenv("QPS_LIMIT"))
	queryLimit, _ := strconv.Atoi(os.Getenv("QUERY_LIMIT"))

	log.Printf("QPS limit :%d\n", qpsLimit)
	log.Printf("Overall query limit :%d\n", queryLimit)

	router.GET("/v1/test-limiter", localmw.LimitPerUser(cmd, qpsLimit, "limit:qps", time.Second*1),
		localmw.LimitPerUser(cmd, queryLimit, "limit:query", 0), func(c *gin.Context) {
			log.Print("Calling /v1/test-limiter")
			c.Status(http.StatusOK)

		})

	if err := router.Run(fmt.Sprintf(":%s", os.Getenv("API_PORT"))); err != nil {
		log.Panic(err)
	}
}
