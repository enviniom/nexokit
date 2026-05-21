package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestLocalLimiterAllowAndRefill(t *testing.T) {
	t.Parallel()

	limiter := NewLocalLimiter(2, 100*time.Millisecond, 50*time.Millisecond)
	t.Cleanup(func() { _ = limiter.Close() })

	ok, err := limiter.Allow(context.Background(), "ip:1.2.3.4")
	if err != nil || !ok {
		t.Fatalf("expected first allow=true,nil, got %v,%v", ok, err)
	}
	ok, _ = limiter.Allow(context.Background(), "ip:1.2.3.4")
	if !ok {
		t.Fatalf("expected second allow=true")
	}
	ok, _ = limiter.Allow(context.Background(), "ip:1.2.3.4")
	if ok {
		t.Fatalf("expected third allow=false (over limit)")
	}

	time.Sleep(120 * time.Millisecond)
	ok, _ = limiter.Allow(context.Background(), "ip:1.2.3.4")
	if !ok {
		t.Fatalf("expected token refill to allow next request")
	}
}

func TestLocalLimiterCleanupRemovesExpiredEntries(t *testing.T) {
	t.Parallel()

	limiter := NewLocalLimiter(1, time.Second, 20*time.Millisecond)
	t.Cleanup(func() { _ = limiter.Close() })

	_, _ = limiter.Allow(context.Background(), "ip:cleanup")
	time.Sleep(70 * time.Millisecond)

	limiter.mu.Lock()
	_, exists := limiter.buckets["ip:cleanup"]
	limiter.mu.Unlock()
	if exists {
		t.Fatalf("expected expired bucket to be removed")
	}
}

func TestLocalLimiterCloseStopsGoroutine(t *testing.T) {
	t.Parallel()

	limiter := NewLocalLimiter(1, time.Second, 10*time.Millisecond)
	if err := limiter.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if err := limiter.Close(); err != nil {
		t.Fatalf("second close should be idempotent: %v", err)
	}
}

func TestRedisLimiterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires redis")
	}

	cfg := RedisLimiterConfig{Addr: "localhost:6379", Limit: 3, Window: 30 * time.Second, DialTimeout: 2 * time.Second, Prefix: "rl"}
	limiter, err := NewRedisLimiter(cfg)
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() { _ = limiter.Close() })

	key := "login:127.0.0.1"
	for i := 0; i < 3; i++ {
		ok, allowErr := limiter.Allow(context.Background(), key)
		if allowErr != nil || !ok {
			t.Fatalf("expected allow on request %d, got %v,%v", i+1, ok, allowErr)
		}
	}
	ok, err := limiter.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected allow error: %v", err)
	}
	if ok {
		t.Fatalf("expected over-limit request denied")
	}

	ttl, err := limiter.client.TTL(context.Background(), "rl:"+key).Result()
	if err != nil {
		t.Fatalf("ttl read failed: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("expected ttl to be set on first request, got %v", ttl)
	}
}

func TestRedisLimiterConnectionErrorPropagation(t *testing.T) {
	limiter := &RedisLimiter{client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 50 * time.Millisecond}), script: redis.NewScript("return {1,1,1}"), prefix: "rl", limit: 1, window: time.Second}
	t.Cleanup(func() { _ = limiter.Close() })

	_, err := limiter.Allow(context.Background(), "global:127.0.0.1")
	if err == nil {
		t.Fatalf("expected redis connection error")
	}
}

type spyLimiter struct {
	allow bool
	err   error
	key   string

	called atomic.Int32
}

func (s *spyLimiter) Allow(_ context.Context, key string) (bool, error) {
	s.key = key
	s.called.Add(1)
	return s.allow, s.err
}
func (s *spyLimiter) Close() error { return nil }

func TestRateLimitMiddlewareHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		enabled    bool
		allow      bool
		err        error
		xff        string
		remoteAddr string
		wantStatus int
		wantCalled int32
		wantKey    string
		wantBody   string
	}{
		{name: "within limit passes", enabled: true, allow: true, remoteAddr: "1.2.3.4:1234", wantStatus: http.StatusOK, wantCalled: 1, wantKey: "login:1.2.3.4"},
		{name: "over limit returns 429", enabled: true, allow: false, remoteAddr: "1.2.3.4:1234", wantStatus: http.StatusTooManyRequests, wantCalled: 1, wantKey: "login:1.2.3.4", wantBody: messages.MsgTooManyRequests},
		{name: "disabled bypasses", enabled: false, allow: false, remoteAddr: "1.2.3.4:1234", wantStatus: http.StatusOK, wantCalled: 0},
		{name: "uses x-forwarded-for first", enabled: true, allow: true, xff: "10.0.0.1, 10.0.0.2", remoteAddr: "1.2.3.4:1234", wantStatus: http.StatusOK, wantCalled: 1, wantKey: "login:10.0.0.1"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			spy := &spyLimiter{allow: tt.allow, err: tt.err}
			r := gin.New()
			r.Use(RateLimitMiddleware(spy, tt.enabled, "login", 5, time.Minute))
			r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/x", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if spy.called.Load() != tt.wantCalled {
				t.Fatalf("expected Allow called %d times, got %d", tt.wantCalled, spy.called.Load())
			}
			if tt.wantKey != "" && spy.key != tt.wantKey {
				t.Fatalf("expected key %q, got %q", tt.wantKey, spy.key)
			}
			if tt.wantBody != "" && !contains(w.Body.String(), tt.wantBody) {
				t.Fatalf("expected body to contain %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func contains(s, sub string) bool { return len(sub) == 0 || (len(s) >= len(sub) && stringContains(s, sub)) }

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRateLimitMiddlewareFailOpenOnLimiterError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &spyLimiter{allow: false, err: errors.New("redis down")}
	r := gin.New()
	r.Use(RateLimitMiddleware(spy, true, "login", 5, time.Minute))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "1.1.1.1:123"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected fail-open status 200, got %d", w.Code)
	}
}
