package main

import (
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type CircularQueue struct {
	data        []int64
	head, tail  int
	count, size int
}

func NewQueue(size int) *CircularQueue {
	return &CircularQueue{
		data:  make([]int64, size),
		head:  0,
		tail:  0,
		count: 0,
		size:  size,
	}
}

func (cq *CircularQueue) IsFull() bool {
	return cq.count == cq.size

}
func (cq *CircularQueue) IsEmpty() bool {
	return cq.count == 0
}
func (cq *CircularQueue) Enqueue(i int64) {
	cq.data[cq.tail] = i
	cq.count++
	cq.tail = (cq.tail + 1) % cq.size
}
func (cq *CircularQueue) Peek() int64 {
	return cq.data[cq.head]
}
func (cq *CircularQueue) Dequeue() int64 {
	v := cq.data[cq.head]
	cq.count--
	cq.head = (cq.head + 1) % cq.size
	return v
}

type RateLimiter struct {
	mu     sync.Mutex
	second int64
	queue  *CircularQueue
}

func NewRateLimiter(limit int, second int64) *RateLimiter {
	return &RateLimiter{
		queue:  NewQueue(limit),
		second: second,
	}
}
func (rl *RateLimiter) expire(now time.Time) {
	startTime := now.Unix() - rl.second
	for !rl.queue.IsEmpty() {
		v := rl.queue.Peek()
		if v >= startTime {
			break
		}
		rl.queue.Dequeue()
	}

}
func (rl *RateLimiter) Enter() {
	now := time.Now()
	rl.expire(now)
	if rl.queue.IsFull() {
		return
	}
	rl.queue.Enqueue(now.Unix())
}

func (rl *RateLimiter) Limit() bool {
	now := time.Now()
	rl.expire(now)
	if rl.queue.IsFull() {
		return true
	}
	return false
}

func main() {
	limiter := NewRateLimiter(3, 5)
	counter := &atomic.Int32{}
	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		slog.Info("request ", slog.Int("counter", int(counter.Load())))
		if limiter.Limit() {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		limiter.Enter()
		w.Header().Set("Content-Type", "text/plain")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})
	err := http.ListenAndServe(":8080", http.DefaultServeMux)
	if err != nil {
		return
	}
}
