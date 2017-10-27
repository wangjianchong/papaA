/*
   copy it from blaccount。
   本文件主要用来进行申请限制（例如限制某一IP在规定时间内可以提交多少次申请）,防止接口被人滥用.
   这是一个桶状模型，桶内的水表示我们允许的接口调用次数，水越多，可调用的次数就越多。
   加水表示桶有了新的处理能力，放水表示桶处理了申请，接口可调用次数减少。
   初始时，桶是满的；当桶内的水不足以满足放水申请时则不能再处理申请。
*/
package main

import (
	"errors"
	// "fmt"
	"sync"
	"time"
)

//允许在TokenInterval时间内(单位为秒)调用TokenQuota次
var TokenQuota int = 4
var TokenInterval int = 1
var ErrorsNotEnough = errors.New("Not enough") //桶内水不足

const TOKEN_GRANULARITY = 1000

type MAP struct {
	lock   sync.RWMutex
	bucket map[string]*tokenBucket
}

func NewMAP() *MAP {
	return &MAP{bucket: make(map[string]*tokenBucket)}
}

func (m *MAP) Set(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.bucket[k]; !ok {
		tb := new(tokenBucket)
		tb.Init(TokenQuota, TokenInterval)
		m.bucket[k] = tb
	}
	return
}

type tokenBucket struct {
	lastQuery time.Time
	tokens    int //桶内剩余的水量
	burst     int //桶的容量
	step      int //每次处理桶流出的水量
	add       int //向桶内加水的速度
	mu        sync.Mutex
}

//初始化一个桶
func (tb *tokenBucket) Init(quota int, interval int) {
	tb.burst = interval * TOKEN_GRANULARITY
	tb.tokens = tb.burst
	tb.step = interval * TOKEN_GRANULARITY / quota
	tb.add = interval * TOKEN_GRANULARITY * quota / interval
	tb.lastQuery = time.Now()
}

//判断桶的水量，水量不足（ErrorsNotEnough）表示桶已经没有处理能力
func (tb *tokenBucket) TokenBucketQuery() error {
	now := time.Now()
	diff := now.Sub(tb.lastQuery)

	token := int(diff.Nanoseconds()/1000000000) * tb.add //上次处理到本次处理之间需要增加的水量
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if token != 0 {
		tb.lastQuery = now
		tb.tokens += token //桶内加水
	}
	if tb.tokens > tb.burst { //桶是否溢出
		tb.tokens = tb.burst
	}
	if tb.tokens >= tb.step {
		tb.tokens -= tb.step
		return nil
	}

	return ErrorsNotEnough
}

// func tokenQuery() {

// 	tb.Init(4, 1)
// 	cnt := 0
// 	for {
// 		err := tb.TokenBucketQuery()
// 		if err != nil {
// 			fmt.Println(err)
// 		} else {
// 			fmt.Println("take")
// 		}
// 		cnt += 1
// 		fmt.Println(cnt)
// 		time.Sleep(200 * time.Millisecond)
// 	}
// }
