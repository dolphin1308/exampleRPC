package transport

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

// HTTPTransport 实现基于HTTP的传输层
type HTTPTransport struct {
	server   *http.Server
	listener net.Listener
	addr     string
	path     string
	mu       sync.Mutex
	conns    chan *HTTPConn
}

// HTTPConn 表示一个HTTP连接
type HTTPConn struct {
	data []byte
	res  chan []byte
	err  chan error
}

// Listen 在指定地址上监听HTTP连接
func (t *HTTPTransport) Listen(addr string) error {
	t.addr = addr
	t.path = "/rpc"
	t.conns = make(chan *HTTPConn, 10) // 缓冲通道，存储连接

	mux := http.NewServeMux()
	mux.HandleFunc(t.path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 创建新连接
		conn := &HTTPConn{
			data: body,
			res:  make(chan []byte),
			err:  make(chan error),
		}

		// 将连接放入通道
		t.conns <- conn

		// 等待响应
		select {
		case resp := <-conn.res:
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		case err := <-conn.err:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	// 监听统计接口
	mux.HandleFunc("/debug/rpc/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("RPC Stats: Not implemented yet"))
	})

	// 监听服务列表接口
	mux.HandleFunc("/debug/rpc/services", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("RPC Services: Not implemented yet"))
	})

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.server = &http.Server{Handler: mux}
	t.listener = listener

	// 启动HTTP服务器
	go t.server.Serve(listener)

	return nil
}

// Accept 接受一个新的HTTP连接
func (t *HTTPTransport) Accept() (Conn, error) {
	if t.conns == nil {
		return nil, errors.New("transport not listening")
	}
	// 从通道中获取一个连接
	conn := <-t.conns
	return conn, nil
}

// Dial 连接到指定地址的HTTP服务器
func (t *HTTPTransport) Dial(addr string) (Conn, error) {
	client := NewHTTPClient(addr)
	return &HTTPClientConn{client: client}, nil
}

// Close 关闭HTTP服务器
func (t *HTTPTransport) Close() error {
	if t.server != nil {
		return t.server.Close()
	}
	return nil
}

// Read 从HTTP请求中读取数据
func (c *HTTPConn) Read() ([]byte, error) {
	return c.data, nil
}

// Write 将数据写入HTTP响应
func (c *HTTPConn) Write(data []byte) error {
	c.res <- data
	return nil
}

// Close 关闭HTTP连接
func (c *HTTPConn) Close() error {
	close(c.res)
	close(c.err)
	return nil
}

// HTTPClient HTTP客户端实现
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(addr string) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{},
		baseURL: "http://" + addr + "/rpc",
	}
}

// Call 发送HTTP请求
func (c *HTTPClient) Call(data []byte) ([]byte, error) {
	resp, err := c.client.Post(c.baseURL, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP error: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// HTTPClientConn 是HTTP客户端的连接
type HTTPClientConn struct {
	client *HTTPClient
	data   []byte
}

// Read 从HTTP服务器读取数据
func (c *HTTPClientConn) Read() ([]byte, error) {
	return c.data, nil
}

// Write 将数据写入HTTP服务器并获取响应
func (c *HTTPClientConn) Write(data []byte) error {
	resp, err := c.client.Call(data)
	if err != nil {
		return err
	}
	c.data = resp
	return nil
}

// Close 关闭HTTP连接
func (c *HTTPClientConn) Close() error {
	return nil
}
