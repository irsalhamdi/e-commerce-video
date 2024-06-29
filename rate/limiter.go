package rate

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	Expiry   int
	Burst    int
	LimitRPS float64
	clients  map[string]*clientLimiter
	mu       sync.RWMutex
}

type clientLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

func NewLimiter(burst int, expiry int, limitRPS float64) *Limiter {
	clients := make(map[string]*clientLimiter)
	lm := &Limiter{
		Expiry:   expiry,
		LimitRPS: limitRPS,
		Burst:    burst,
		clients:  clients,
	}
	go lm.refresh()
	return lm
}

func (l *Limiter) Check(id string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cl, ok := l.clients[id]
	if !ok {
		l.clients[id] = &clientLimiter{
			limiter:    rate.NewLimiter(rate.Limit(l.LimitRPS), l.Burst),
			lastAccess: time.Now(),
		}
		return l.clients[id].limiter.Allow()
	}
	cl.lastAccess = time.Now()
	return cl.limiter.Allow()
}

func (l *Limiter) refresh() {
	for {
		time.Sleep(time.Minute)

		l.mu.Lock()
		for id, v := range l.clients {
			if time.Since(v.lastAccess) > time.Duration(l.Expiry)*time.Minute {
				delete(l.clients, id)
			}
		}
		l.mu.Unlock()
	}
}

func Every(interval time.Duration) float64 {
	return float64(rate.Every(interval))
}
