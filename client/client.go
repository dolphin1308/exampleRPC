package client

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"rpc/codec"
	"rpc/protocol"
	"rpc/transport"
)

// Client RPC客户端
type Client struct {
	transport   transport.Transport // 传输层
	conn        transport.Conn      // 当前连接
	serverAddr  string              // 服务器地址
	codecType   codec.Type          // 编解码类型
	serializer  codec.Codec         // 序列化工具
	mu          sync.Mutex          // 保护连接
	isConnected bool                // 是否已连接
}

// Option 配置选项
type Option struct {
	TransportType transport.TransportType // 传输类型
	CodecType     codec.Type              // 编解码类型
	Timeout       time.Duration           // 超时时间
}

// DefaultOption 默认配置
var DefaultOption = &Option{
	TransportType: transport.TCP,
	CodecType:     codec.JSON,
	Timeout:       time.Second * 10,
}

// NewClient 创建客户端实例
func NewClient(addr string, opt *Option) *Client {
	if opt == nil {
		opt = DefaultOption
	}

	c := &Client{
		serverAddr: addr,
		codecType:  opt.CodecType,
		transport:  transport.NewTransport(opt.TransportType),
		serializer: codec.NewCodec(opt.CodecType),
	}

	return c
}

// Connect 连接到服务器
func (client *Client) Connect() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.isConnected {
		return nil
	}

	conn, err := client.transport.Dial(client.serverAddr)
	if err != nil {
		return err
	}

	client.conn = conn
	client.isConnected = true
	return nil
}

// Close 关闭连接
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if !client.isConnected || client.conn == nil {
		return nil
	}

	err := client.conn.Close()
	client.isConnected = false
	return err
}

// Call 远程调用方法
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	// 确保连接已建立
	if err := client.Connect(); err != nil {
		return err
	}

	// 分割服务名和方法名
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		return errors.New("service/method request ill-formed: " + serviceMethod)
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]

	// 序列化参数
	argBytes, err := client.serializer.Encode(args)
	if err != nil {
		return fmt.Errorf("encode arguments error: %v", err)
	}

	// 构造请求
	header := &protocol.Header{
		MagicNumber:   protocol.MagicNumber,
		Version:       protocol.Version,
		MessageType:   protocol.Request,
		SerializeType: byte(client.codecType),
		ServiceLength: uint16(len(serviceName)),
		MethodLength:  uint16(len(methodName)),
		PayloadLength: uint32(len(argBytes)),
	}

	// 编码请求头
	headerBytes := protocol.EncodeHeader(header)

	// 合并请求数据
	reqData := append(headerBytes, []byte(serviceName)...)
	reqData = append(reqData, []byte(methodName)...)
	reqData = append(reqData, argBytes...)

	client.mu.Lock()
	defer client.mu.Unlock()

	// 发送请求
	if err := client.conn.Write(reqData); err != nil {
		client.isConnected = false
		return fmt.Errorf("send request error: %v", err)
	}

	// 接收响应
	respData, err := client.conn.Read()
	if err != nil {
		client.isConnected = false
		return fmt.Errorf("read response error: %v", err)
	}

	// 解析响应头
	if len(respData) < protocol.HeaderSize {
		return errors.New("invalid response header")
	}

	respHeader, err := protocol.DecodeHeader(respData[:protocol.HeaderSize])
	if err != nil {
		return err
	}

	// 检查响应类型
	if respHeader.MessageType != protocol.Response {
		return errors.New("invalid message type in response")
	}

	// 解析响应数据
	if len(respData) < protocol.HeaderSize+int(respHeader.PayloadLength) {
		return errors.New("invalid response data")
	}

	// 解码响应
	var response protocol.ResponseMessage
	if err := client.serializer.Decode(respData[protocol.HeaderSize:protocol.HeaderSize+respHeader.PayloadLength], &response); err != nil {
		return fmt.Errorf("decode response error: %v", err)
	}

	// 检查响应中是否有错误
	if response.Error != "" {
		return errors.New(response.Error)
	}

	// 将结果解码到reply中
	if response.Result != nil {
		// 将response.Result赋值给reply
		respData, err := client.serializer.Encode(response.Result)
		if err != nil {
			return fmt.Errorf("encode result error: %v", err)
		}
		return client.serializer.Decode(respData, reply)
	}

	return nil
}
