package handler

import (
	"sync"
	"time"
)

// loginLimiter 基于 IP 的登录频率限制
// 每个 IP 每分钟最多尝试 5 次，超过后返回 429

type ipEntry struct {
	count    int
	windowStart time.Time
}

type loginRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
}

func newLoginRateLimiter() *loginRateLimiter {
	l := &loginRateLimiter{entries: make(map[string]*ipEntry)}
	go l.cleanup()
	return l
}

// Allow 返回 true 表示允许本次请求；false 表示已超限
func (l *loginRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.entries[ip]
	if !ok || now.Sub(e.windowStart) > time.Minute {
		// 新窗口
		l.entries[ip] = &ipEntry{count: 1, windowStart: now}
		return true
	}
	e.count++
	return e.count <= 5
}

// cleanup 每 10 分钟清理超过 30 分钟未活动的条目
func (l *loginRateLimiter) cleanup() {
	for range time.Tick(10 * time.Minute) {
		l.mu.Lock()
		for ip, e := range l.entries {
			if time.Since(e.windowStart) > 30*time.Minute {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}
