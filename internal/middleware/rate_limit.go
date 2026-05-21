package middleware

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const defaultCleanupInterval = 5 * time.Minute

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Close() error
}

type NoopLimiter struct{}

func NewNoopLimiter() *NoopLimiter { return &NoopLimiter{} }

func (l *NoopLimiter) Allow(context.Context, string) (bool, error) { return true, nil }
func (l *NoopLimiter) Close() error                                { return nil }

type localBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type LocalLimiter struct {
	rate            rate.Limit
	burst           int
	cleanupInterval time.Duration

	mu      sync.Mutex
	buckets map[string]*localBucket

	stopCh chan struct{}
	once   sync.Once
}

func NewLocalLimiter(limit int, window, cleanupInterval time.Duration) *LocalLimiter {
	if window <= 0 {
		window = time.Minute
	}
	if limit <= 0 {
		return &LocalLimiter{rate: rate.Inf, burst: 1, buckets: map[string]*localBucket{}, cleanupInterval: defaultCleanupInterval, stopCh: make(chan struct{})}
	}
	if cleanupInterval <= 0 {
		cleanupInterval = defaultCleanupInterval
	}
	l := &LocalLimiter{
		rate:            rate.Every(window / time.Duration(limit)),
		burst:           limit,
		cleanupInterval: cleanupInterval,
		buckets:         map[string]*localBucket{},
		stopCh:          make(chan struct{}),
	}
	go l.runCleanup()
	return l
}

func (l *LocalLimiter) Allow(_ context.Context, key string) (bool, error) {
	now := time.Now()
	l.mu.Lock()
	bucket, ok := l.buckets[key]
	if !ok {
		bucket = &localBucket{limiter: rate.NewLimiter(l.rate, l.burst), lastSeen: now}
		l.buckets[key] = bucket
	}
	bucket.lastSeen = now
	allowed := bucket.limiter.Allow()
	l.mu.Unlock()
	return allowed, nil
}

func (l *LocalLimiter) Close() error {
	l.once.Do(func() { close(l.stopCh) })
	return nil
}

func (l *LocalLimiter) runCleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanupExpired()
		case <-l.stopCh:
			return
		}
	}
}

func (l *LocalLimiter) cleanupExpired() {
	now := time.Now()
	l.mu.Lock()
	for key, bucket := range l.buckets {
		if now.Sub(bucket.lastSeen) > l.cleanupInterval {
			delete(l.buckets, key)
		}
	}
	l.mu.Unlock()
}

type RedisLimiter struct {
	client *redis.Client
	script *redis.Script
	prefix string
	limit  int
	window time.Duration
}

type RedisLimiterConfig struct {
	Addr        string
	Password    string
	DB          int
	DialTimeout time.Duration
	Prefix      string
	Limit       int
	Window      time.Duration
}

func NewRedisLimiter(cfg RedisLimiterConfig) (*RedisLimiter, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB, DialTimeout: cfg.DialTimeout})
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis limiter: %w", err)
	}
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "rl"
	}
	return &RedisLimiter{client: client, script: redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[2])
end
local ttl = redis.call("TTL", KEYS[1])
local allowed = 0
if tonumber(current) <= tonumber(ARGV[1]) then
  allowed = 1
end
return {allowed, current, ttl}
`), prefix: prefix, limit: cfg.Limit, window: cfg.Window}, nil
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
	res, err := l.script.Run(ctx, l.client, []string{redisKey}, l.limit, int(l.window.Seconds())).Result()
	if err != nil {
		return false, err
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) != 3 {
		return false, fmt.Errorf("unexpected redis limiter response")
	}
	allowed, ok := arr[0].(int64)
	if !ok {
		return false, fmt.Errorf("invalid allowed type")
	}
	return allowed == 1, nil
}

func (l *RedisLimiter) Close() error {
	return l.client.Close()
}

func RateLimitMiddleware(limiter Limiter, enabled bool, scope string, limit int, window time.Duration) gin.HandlerFunc {
	_ = limit
	_ = window
	if limiter == nil {
		limiter = NewNoopLimiter()
	}
	if scope == "" {
		scope = "global"
	}
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		ip := clientIP(c)
		allowed, err := limiter.Allow(c.Request.Context(), fmt.Sprintf("%s:%s", scope, ip))
		if err != nil {
			c.Next()
			return
		}
		if !allowed {
			response.TooManyRequests(c, messages.MsgTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return c.Request.RemoteAddr
}
