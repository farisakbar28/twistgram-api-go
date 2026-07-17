package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*clientVisitor)
	mtx      sync.Mutex
)

func init() {
	go cleanupVisitors()
}

func getVisitor(ip string, r rate.Limit, b int) *rate.Limiter {
	mtx.Lock()
	defer mtx.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// Proteksi dari memory leak bila diserang jutaan bot/IP spoofing
		if len(visitors) > 10000 {
			// Bersihkan paksa map secara asinkron jika terlalu penuh
			go func() {
				mtx.Lock()
				for k := range visitors { delete(visitors, k) }
				mtx.Unlock()
			}()
		}
		limiter := rate.NewLimiter(r, b)
		visitors[ip] = &clientVisitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)
		mtx.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mtx.Unlock()
	}
}

// RateLimit adalah middleware untuk membatasi jumlah request per IP.
// r: limit per detik (contoh: 5 untuk 5 req/sec).
// b: burst limit (maksimal request dadakan dalam seketika).
func RateLimit(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip, r, b)

		// Set rate limit info headers
		c.Header("X-RateLimit-Limit", "5")

		remaining := limiter.Tokens()
		c.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))

		if !limiter.Allow() {
			// Estimate wait time: token refill rate is r tokens/sec
			retryAfter := int(1.0 / float64(r)) + 1 // at least 1 second
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
