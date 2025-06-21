package transport

import (
	"encoding/binary"
	"errors"
	"net"
)

// UDPTransport 实现基于UDP的传输层
type UDPTransport struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

// UDPConn 表示一个UDP连接(伪连接，含有远程地址连接)
type UDPConn struct {
	conn  *net.UDPConn
	raddr *net.UDPAddr
}

// ListenUDP UDP监听
func (t *UDPTransport) ListenUDP(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	t.conn = conn
	t.addr = udpAddr
	return nil
}

// Accept 接受一个UDP“连接”（其实是收到数据时的远程地址）
func (t *UDPTransport) Accept() (Conn, error) {
	if t.conn == nil {
		return nil, errors.New("UDPTransport not listening")
	}

	// 先读取数据长度前缀
	buf := make([]byte, 4096)
	n, addr, err := t.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	c := &UDPConn{
		conn:  t.conn,
		raddr: addr,
	}
	return c, nil
}

// DialUDP 建立UDP连接
func (t *UDPTransport) Dial(addr string) (Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPConn{conn: conn, raddr: udpAddr}, nil
}

// Close UDP监听关闭
func (t *UDPTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// Write 向指定地址写入数据，带有四字节前缀
func (c *UDPConn) Write(data []byte) error {
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(data)))

	msg := append(sizeBuf, data...)
	_, err := c.conn.WriteToUDP(msg, c.raddr)
	return err
}

// Read 读取UDP数据（含长度前缀）
func (c *UDPConn) Read() ([]byte, error) {
	buf := make([]byte, 4096)
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	if n < 4 {
		return nil, errors.New("invalid data length")
	}
	size := binary.BigEndian.Uint32(buf[:4])
	if int(size) > n-4 {
		return nil, errors.New("incomplete data")
	}
	return buf[4 : 4+size], nil
}

// Close 关闭udp连接
func (c *UDPConn) Close() error {
	return c.Close()
}
