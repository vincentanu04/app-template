package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	Requests    int
	WindowStart time.Time
}

var (
	mu       sync.Mutex
	clients  = make(map[string]*client)
	rate     = 100 // max requests per window
	interval = time.Minute
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		mu.Lock()
		c, exists := clients[ip]
		if !exists || time.Since(c.WindowStart) > interval {
			c = &client{Requests: 1, WindowStart: time.Now()}
			clients[ip] = c
		} else {
			c.Requests++
		}
		reqs := c.Requests
		mu.Unlock()

		if reqs > rate {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
