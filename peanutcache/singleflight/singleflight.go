// Copyright 2021 Peanutzhen. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package singleflight

import (
	"sync"
)

// singleflight 为peanutcache提供缓存击穿的保护
// 当cache并发访问peer获取缓存时 如果peer未缓存该值
// 则会向db发送大量的请求获取 造成db的压力骤增
// 因此 将所有由key产生的请求抽象成flight
// 这个flight只会起飞一次(single) 这样就可以缓存击穿的可能性
// flight载有我们要的缓存数据 称为packet

// sigleFlight的原理：
// 在多个并发请求出发的回调操作中，只有第一个回调方法被执行，
// 其余请求堵塞等待第一个回调函数执行完成后直接取结果
// 以此保证同一时刻只有一个回调方法执行，达到防止缓存击穿的目的

// 正在进行中或者已经结束的请求
type packet struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// 管理不同key的请求
type Flight struct { 
	mu     sync.Mutex
	flight map[string]*packet
}

// Fly 负责key航班的飞行 fn是获取packet的方法
// WaitGroup Add用于增加goroutine数量，Done用于减少goroutine数量，Wait用于等待所有需要等待的goroutine
func (f *Flight) Fly(key string, fn func() (interface{}, error)) (interface{}, error) {
	f.mu.Lock()
	if f.flight == nil {
		f.flight = make(map[string]*packet)
	}
	if p, ok := f.flight[key]; ok { // 如果该key请求正在进行，则等待
		f.mu.Unlock()
		p.wg.Wait() // 等待协程结束
		return p.val, p.err // 返回目标值
	}
	p := new(packet)
	p.wg.Add(1) // 发起请求前加锁
	f.flight[key] = p // 表明该key已经有请求在进行
	f.mu.Unlock()

	p.val, p.err = fn() // 执行请求
	p.wg.Done()

	f.mu.Lock()
	delete(f.flight, key) // 航班已完成
	f.mu.Unlock()

	return p.val, p.err
}
