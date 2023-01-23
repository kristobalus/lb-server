package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
)

var (
	ctx context.Context
	rdb *redis.Client
	r   *gin.Engine
)

func init() {

	ctx = context.Background()

	REDIS_SENTINELS, isSentinel := os.LookupEnv("REDIS_SENTINELS")
	if isSentinel {
		sentinels := strings.Split(REDIS_SENTINELS, ",")
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    os.Getenv("REDIS_MASTER_NAME"),
			SentinelAddrs: sentinels,
		})
	} else {

		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	}

	_, isRelease := os.LookupEnv("RELEASE")
	if isRelease {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	} else {
		r = gin.Default()
	}
}

func main() {

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/get", func(c *gin.Context) {

		orgId := c.Query("orgId")
		eventId := c.Query("eventId")
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "0"), 10, 32)
		pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "20"), 10, 32)

		users, count := getMainLeaderboard(orgId, eventId, page, pageSize)

		c.JSON(http.StatusOK, &LbResponse{
			Meta: LbMeta{
				Page:     page,
				PageSize: pageSize,
				Count:    count,
				Type:     "main",
			},
			Data: users,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
