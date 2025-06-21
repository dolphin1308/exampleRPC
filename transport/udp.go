package transport

import "net"

// UDPTransport 实现基于UDP的传输层
type UDPTransport struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

// UDPConn 表示一个UDP连接
