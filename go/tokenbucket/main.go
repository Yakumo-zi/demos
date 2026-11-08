package main

import (
	"fmt"
	"time"
)

type TokenBucket struct {
	capacity    float64 // 最大容量
	ratio       float64 // 单位时间内每秒恢复的 token
	lastUpdated time.Time
	count       float64 // 还剩多少 token
}

func NewTokenBucket(capacity, ratio float64) *TokenBucket {
	return &TokenBucket{
		capacity:    capacity,
		ratio:       ratio,
		lastUpdated: time.Now(),
		count:       capacity,
	}
}

func (tb *TokenBucket) Require() bool {
	now := time.Now()
	diff := now.Sub(tb.lastUpdated).Seconds()
	tb.lastUpdated = now
	//惰性补充 token，用每秒速率*上次更新距离当前时间的秒数，然后比较剩余 token 数 + 补充的 token 数和最大容量
	//取最小值
	tokens := tb.ratio * diff
	tb.count = min(tb.capacity, tokens+tb.count)
	fmt.Printf("count:%f\n", tb.count)
	if tb.count < 1 {
		return false
	}
	tb.count--
	return false
}

func main() {
	tb := NewTokenBucket(10, 1)
	for range 13 {
		time.Sleep(time.Millisecond * 200)
		fmt.Println(tb.Require())
	}
}
