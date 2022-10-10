package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// token bucket algorithm

type TokenBucket struct {
	Rate           time.Duration // 1 token per rate
	maxTokens      int64
	currentTokens  int64
	lock           sync.Mutex
	lastRefilledAt time.Time
}

type RateLimiter interface {
	TryRequest() bool
}

type TokenRateLimiter struct {
	bucket *TokenBucket
}

// ==

func NewRateLimiter(maxTokens int64, rate time.Duration) RateLimiter {
	return &TokenRateLimiter{
		bucket: &TokenBucket{
			Rate:           rate,
			maxTokens:      maxTokens,
			currentTokens:  maxTokens,
			lastRefilledAt: time.Now(),
		},
	}
}

func (rl *TokenRateLimiter) TryRequest() bool {
	rl.bucket.lock.Lock()
	defer rl.bucket.lock.Unlock()
	rl.tryRefill()

	if rl.bucket.currentTokens > 0 {
		rl.bucket.currentTokens--
		return true
	}

	return false
}

func (rl *TokenRateLimiter) tryRefill() {
	since := time.Since(rl.bucket.lastRefilledAt)

	if since < rl.bucket.Rate {
		return
	}
	newTokensCount := rl.bucket.currentTokens + (since.Nanoseconds() / rl.bucket.Rate.Nanoseconds())
	maxTokensCount := rl.bucket.maxTokens
	if newTokensCount < maxTokensCount {
		rl.bucket.currentTokens = newTokensCount
	} else {
		rl.bucket.currentTokens = maxTokensCount
	}

	fmt.Printf("Tokens filled! current Token : %d\n", rl.bucket.currentTokens)

	rl.bucket.lastRefilledAt = time.Now()
}

// ==

var rateLimiter = NewRateLimiter(10, 5*time.Second)

func main() {
	http.HandleFunc("/", request)
	http.ListenAndServe(":7070", nil)
}

func request(http.ResponseWriter, *http.Request) {
	if !rateLimiter.TryRequest() {
		fmt.Println("Rate limit exceeded!")
		return
	}

	fmt.Println("Request success!")
}
