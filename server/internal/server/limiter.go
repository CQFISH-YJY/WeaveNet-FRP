package server

// 令牌桶限速器：按用户 Token 维度限速。

import (
	"io"
	"sync"
	"time"
)

// TokenBucket 单个用户的令牌桶。
type TokenBucket struct {
	rate       float64 // 每秒补充令牌数（字节）
	capacity   float64 // 桶容量（突发）
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// newTokenBucket 创建令牌桶，rateBytes 为每秒字节数。
func newTokenBucket(rateBytes float64) *TokenBucket {
	capacity := rateBytes // 突发容量 = 1 秒流量
	if capacity < 1024 {
		capacity = 1024
	}
	return &TokenBucket{
		rate:       rateBytes,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// take 尝试取 n 个令牌，返回实际可取的字节数（阻塞直到足够）。
func (b *TokenBucket) take(n int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}
	if b.tokens < float64(n) {
		// 阻塞等待令牌
		need := float64(n) - b.tokens
		wait := time.Duration(need / b.rate * float64(time.Second))
		if wait > time.Second {
			wait = time.Second
		}
		time.Sleep(wait)
		now = time.Now()
		elapsed = now.Sub(b.lastRefill).Seconds()
		b.tokens += elapsed * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}
	b.tokens -= float64(n)
	return n
}

// RateLimiter 管理所有用户的令牌桶。
type RateLimiter struct {
	mu     sync.RWMutex
	limits map[string]*TokenBucket
}

// NewRateLimiter 创建限速器。
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{limits: make(map[string]*TokenBucket)}
}

// SetLimit 设置用户限速（Kbps）。kbps<=0 表示不限制；-1 表示封禁。
func (r *RateLimiter) SetLimit(userToken string, kbps int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if kbps <= 0 {
		delete(r.limits, userToken)
		return
	}
	rateBytes := float64(kbps) * 125 // Kbps -> bytes/s
	r.limits[userToken] = newTokenBucket(rateBytes)
}

// Get 获取用户限速状态。ok=false 表示未注册；-1 表示封禁。
func (r *RateLimiter) Get(userToken string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.limits[userToken]
	if !ok {
		return 0, false
	}
	return int(b.rate / 125), true
}

// Writer 返回限速写入器。
func (r *RateLimiter) Writer(w io.Writer, userToken string) io.Writer {
	r.mu.RLock()
	bucket := r.limits[userToken]
	r.mu.RUnlock()
	if bucket == nil {
		return w
	}
	return &rateWriter{w: w, bucket: bucket}
}

// rateWriter 限速写入器。
type rateWriter struct {
	w      io.Writer
	bucket *TokenBucket
}

func (rw *rateWriter) Write(p []byte) (int, error) {
	written := 0
	for written < len(p) {
		chunk := len(p) - written
		if chunk > 64*1024 {
			chunk = 64 * 1024
		}
		rw.bucket.take(chunk)
		n, err := rw.w.Write(p[written : written+chunk])
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
