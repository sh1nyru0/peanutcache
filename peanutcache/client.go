// Copyright 2021 Peanutzhen. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package peanutcache

import (
	"context"
	"fmt"
	pb "peanutcache/peanutcachepb"
	"peanutcache/registry"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// client 模块实现peanutcache访问其他远程节点 从而获取缓存的能力

type client struct {
	name string // 服务名称 pcache/ip:addr
}

// Fetch 从remote peer获取对应缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {
	// 创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	// 发现服务 取得与服务的连接
	conn, err := registry.EtcdDial(cli, c.name)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	grpcClient := pb.NewPeanutCacheClient(conn) // 与PeanucCache服务交互的接口
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 创建了一个10s超时的上下文
	defer cancel()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.name)
	}

	return resp.GetValue(), nil
}

func NewClient(service string) *client {
	return &client{name: service}
}

// 测试Client是否实现了Fetcher接口
var _ Fetcher = (*client)(nil)
