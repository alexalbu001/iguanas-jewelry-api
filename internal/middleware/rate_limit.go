package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	limiterredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func NewRateLimiter(rdb *redis.Client, rate string) gin.HandlerFunc {
	store, _ := limiterredis.NewStore(rdb)

	rateLimiter := limiter.New(store, limiter.Rate{
		Period: 30 * time.Minute,
		Limit:  100,
	})
	if rate != "" {
		if parsedRate, err := limiter.NewRateFromFormatted(rate); err == nil {
			rateLimiter = limiter.New(store, parsedRate)
		}
	}
	return mgin.NewMiddleware(rateLimiter)

}
