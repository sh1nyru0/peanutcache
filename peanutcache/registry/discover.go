
package registry

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

// EtcdDial 通过etcd实现gRPC服务发现，并建立一个与目标服务的连接
// 通过提供一个etcd client和service name即可获得Connection
func EtcdDial(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(
		"etcd:///"+service,// 服务地址格式，etcd将处理这个地址来解析服务位置
		grpc.WithResolvers(etcdResolver), // 将之前创建的etcd解析器传递给gRPC，以便gRPC可以使用etcd进行服务发现
		grpc.WithInsecure(),// 不用TLS加密
		grpc.WithBlock(),// Dial直到建立连接将一直阻塞
	)
}
