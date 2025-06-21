package transport

// Transport 定义传输层接口
type Transport interface {
	Listen(addr string) error       // 服务端监听
	Accept() (Conn, error)          // 接受连接
	Dial(addr string) (Conn, error) // 客户端连接
	Close() error                   // 关闭
}

// Conn 定义连接接口
type Conn interface {
	Read() ([]byte, error) // 读取数据
	Write([]byte) error    // 写入数据
	Close() error          // 关闭连接
}

// TransportType 表示传输类型
type TransportType byte

const (
	TCP  TransportType = iota // 0
	HTTP                      // 1
)

// NewTransport 创建传输层实例
func NewTransport(transportType TransportType) Transport {
	switch transportType {
	case TCP:
		return &TCPTransport{}
	case HTTP:
		return &HTTPTransport{}
	default:
		return &TCPTransport{} // 默认使用TCP
	}
}
