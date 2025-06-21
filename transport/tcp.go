package transport

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

// TCPTransport 实现基于TCP的传输层
type TCPTransport struct {
	listener net.Listener
}

// TCPConn 表示一个TCP连接
type TCPConn struct {
	conn net.Conn
}

// Listen 在指定地址上监听TCP连接
func (t *TCPTransport) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	t.listener = listener
	return nil
}

// Accept 接受一个新的TCP连接
func (t *TCPTransport) Accept() (Conn, error) {
	if t.listener == nil {
		return nil, errors.New("transport not listening")
	}
	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &TCPConn{conn: conn}, nil
}

// Dial 连接到指定地址的TCP服务器
func (t *TCPTransport) Dial(addr string) (Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPConn{conn: conn}, nil
}

// Close 关闭TCP监听器
func (t *TCPTransport) Close() error {
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

// Read 从TCP连接中读取数据
func (c *TCPConn) Read() ([]byte, error) {
	// 首先读取数据长度（4字节）
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, sizeBuf); err != nil {
		return nil, err
	}

	// 解析数据长度
	size := binary.BigEndian.Uint32(sizeBuf)

	// 读取实际数据
	data := make([]byte, size)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Write 将数据写入TCP连接
func (c *TCPConn) Write(data []byte) error {
	// 先写入数据长度（4字节）
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(data)))

	// 写入长度
	if _, err := c.conn.Write(sizeBuf); err != nil {
		return err
	}

	// 写入数据
	_, err := c.conn.Write(data)
	return err
}

// Close 关闭TCP连接
func (c *TCPConn) Close() error {
	return c.conn.Close()
}
