package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ── 滑动窗口限速器 ────────────────────────────────────────────────────────────
//
// 相比固定窗口，滑动窗口不存在"边界翻倍"漏洞：
// 任意连续 windowSize 时间段内，请求数都不会超过 maxReqs。

type slidingEntry struct {
	timestamps []time.Time
	mu         sync.Mutex
}

func (e *slidingEntry) allow(maxReqs int, window time.Duration) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// 过滤掉窗口外的旧记录
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = valid

	if len(e.timestamps) >= maxReqs {
		return false
	}
	e.timestamps = append(e.timestamps, now)
	return true
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*slidingEntry
	maxReqs int
	window  time.Duration
}

func newRateLimiter(maxReqs int, window time.Duration) *rateLimiter {
	l := &rateLimiter{
		entries: make(map[string]*slidingEntry),
		maxReqs: maxReqs,
		window:  window,
	}
	go l.cleanup()
	return l
}

func (l *rateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	e, ok := l.entries[ip]
	if !ok {
		e = &slidingEntry{}
		l.entries[ip] = e
	}
	l.mu.Unlock()
	return e.allow(l.maxReqs, l.window)
}

// cleanup 每 5 分钟清理超过 2 个窗口时长未活动的条目
func (l *rateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		cutoff := time.Now().Add(-2 * l.window)
		for ip, e := range l.entries {
			e.mu.Lock()
			allOld := true
			for _, t := range e.timestamps {
				if t.After(cutoff) {
					allOld = false
					break
				}
			}
			e.mu.Unlock()
			if allOld {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// ── 真实 IP 提取（防止 X-Forwarded-For 伪造） ────────────────────────────────
//
// 只有当请求来自受信任的私有网络时，才相信 X-Forwarded-For / X-Real-IP。
// 直连公网的请求直接使用 RemoteAddr，完全不信任任何转发头。

var privateRanges = func() []*net.IPNet {
	var nets []*net.IPNet
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	} {
		_, n, _ := net.ParseCIDR(cidr)
		nets = append(nets, n)
	}
	return nets
}()

func isPrivateIP(ip net.IP) bool {
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

// RealIP 返回请求的真实客户端 IP。
// 仅当直连 IP 属于私有网段（即经过可信反代）时才读取转发头。
func RealIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	remoteIP := net.ParseIP(remoteHost)

	// 直连 IP 不是私有地址 → 公网直连，完全不信任转发头
	if remoteIP == nil || !isPrivateIP(remoteIP) {
		return remoteHost
	}

	// 来自私有网段（反代或本地） → 读取转发头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	return remoteHost
}

// GinRealIP 是 RealIP 的 Gin 适配版
func GinRealIP(c *gin.Context) string {
	return RealIP(c.Request)
}
