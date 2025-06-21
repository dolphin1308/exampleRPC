package server

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"rpc/codec"
	"rpc/protocol"
	"rpc/transport"
)

// methodType 保存方法的信息
type methodType struct {
	method    reflect.Method // 方法本身
	ArgType   reflect.Type   // 第一个参数类型
	ReplyType reflect.Type   // 第二个参数类型（返回值）
}

// service 保存服务的信息
type service struct {
	name    string                 // 服务名称
	typ     reflect.Type           // 服务类型
	rcvr    reflect.Value          // 服务实例
	methods map[string]*methodType // 服务方法集合
}

// Server RPC服务器
type Server struct {
	mu         sync.RWMutex        // 保护services
	services   map[string]*service // 注册的服务
	transport  transport.Transport // 传输层
	serializer codec.Codec         // 序列化工具
}

// NewServer 创建RPC服务器
func NewServer(transportType transport.TransportType, codecType codec.Type) *Server {
	return &Server{
		services:   make(map[string]*service),
		transport:  transport.NewTransport(transportType),
		serializer: codec.NewCodec(codecType),
	}
}

// Register 注册服务
func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if s == nil {
		return errors.New("invalid service")
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if _, exists := server.services[s.name]; exists {
		return errors.New("service already defined: " + s.name)
	}

	server.services[s.name] = s
	return nil
}

// newService 创建服务信息
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.typ = reflect.TypeOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()

	if s.name == "" {
		return nil
	}

	s.methods = make(map[string]*methodType)

	// 遍历方法
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		mtype := method.Type

		// 方法必须是导出的
		if method.PkgPath != "" {
			continue
		}

		// 检查方法签名：func(receiver, args, *reply) error
		if mtype.NumIn() != 3 || mtype.NumOut() != 1 {
			continue
		}

		// 返回值类型必须是error
		if returnType := mtype.Out(0); returnType != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		// 第三个参数必须是指针类型（用于返回值）
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}

		// 记录有效的方法
		s.methods[method.Name] = &methodType{
			method:    method,
			ArgType:   mtype.In(1),
			ReplyType: replyType,
		}
	}

	if len(s.methods) == 0 {
		return nil
	}

	return s
}

// Serve 启动RPC服务
func (server *Server) Serve(addr string) error {
	err := server.transport.Listen(addr)
	if err != nil {
		return err
	}

	log.Printf("RPC Server listening on %s\n", addr)

	for {
		conn, err := server.transport.Accept()
		if err != nil {
			log.Printf("RPC server accept error: %v\n", err)
			continue
		}

		go server.handleConn(conn)
	}
}

// handleConn 处理连接请求
func (server *Server) handleConn(conn transport.Conn) {
	defer conn.Close()

	for {
		// 读取请求数据
		data, err := conn.Read()
		if err != nil {
			log.Printf("Read error: %v\n", err)
			return
		}

		// 解析请求头
		if len(data) < protocol.HeaderSize {
			log.Println("Invalid request: header too small")
			continue
		}

		header, err := protocol.DecodeHeader(data[:protocol.HeaderSize])
		if err != nil {
			log.Printf("Decode header error: %v\n", err)
			continue
		}

		// 提取服务名和方法名
		serviceNameEnd := protocol.HeaderSize + header.ServiceLength
		methodNameEnd := serviceNameEnd + header.MethodLength

		if len(data) < int(methodNameEnd) {
			log.Println("Invalid request: data too small")
			continue
		}

		serviceName := string(data[protocol.HeaderSize:serviceNameEnd])
		methodName := string(data[serviceNameEnd:methodNameEnd])

		// 提取参数数据
		payloadStart := methodNameEnd
		payloadEnd := uint32(payloadStart) + header.PayloadLength

		if uint32(len(data)) < payloadEnd {
			log.Println("Invalid request: payload too small")
			continue
		}
		payload := data[payloadStart:payloadEnd]

		// 调用服务方法
		resp, err := server.call(serviceName, methodName, payload)
		if err != nil {
			log.Printf("Call error: %v\n", err)
			// 发送错误响应
			errorResp := &protocol.ResponseMessage{Error: err.Error()}
			respData, _ := server.serializer.Encode(errorResp)

			header := &protocol.Header{
				MagicNumber:   protocol.MagicNumber,
				Version:       protocol.Version,
				MessageType:   protocol.Response,
				SerializeType: header.SerializeType,
				PayloadLength: uint32(len(respData)),
			}

			headerBytes := protocol.EncodeHeader(header)
			conn.Write(append(headerBytes, respData...))
			continue
		}

		// 发送成功响应
		if err := conn.Write(resp); err != nil {
			log.Printf("Write error: %v\n", err)
			return
		}
	}
}

// call 调用服务方法
func (server *Server) call(serviceName, methodName string, argBytes []byte) ([]byte, error) {
	server.mu.RLock()
	service, ok := server.services[serviceName]
	server.mu.RUnlock()

	if !ok {
		return nil, errors.New("service not found: " + serviceName)
	}

	mtype, ok := service.methods[methodName]
	if !ok {
		return nil, errors.New("method not found: " + methodName)
	}

	// 创建参数实例
	argv := reflect.New(mtype.ArgType)
	replyv := reflect.New(mtype.ReplyType.Elem())

	// 解析参数
	if err := server.serializer.Decode(argBytes, argv.Interface()); err != nil {
		return nil, fmt.Errorf("decode argument error: %v", err)
	}

	// 调用方法
	function := mtype.method.Func
	returnValues := function.Call([]reflect.Value{service.rcvr, argv.Elem(), replyv})

	// 处理错误
	errInter := returnValues[0].Interface()
	if errInter != nil {
		err := errInter.(error)
		return nil, err
	}

	// 构造响应
	response := &protocol.ResponseMessage{Result: replyv.Interface()}
	respBytes, err := server.serializer.Encode(response)
	if err != nil {
		return nil, fmt.Errorf("encode response error: %v", err)
	}

	// 构造响应头
	header := &protocol.Header{
		MagicNumber:   protocol.MagicNumber,
		Version:       protocol.Version,
		MessageType:   protocol.Response,
		PayloadLength: uint32(len(respBytes)),
	}

	headerBytes := protocol.EncodeHeader(header)
	return append(headerBytes, respBytes...), nil
}

// findMethod 解析服务方法
func (server *Server) findMethod(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("service/method request ill-formed: " + serviceMethod)
		return
	}

	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]

	server.mu.RLock()
	svc = server.services[serviceName]
	server.mu.RUnlock()
	if svc == nil {
		err = errors.New("can't find service " + serviceName)
		return
	}

	mtype = svc.methods[methodName]
	if mtype == nil {
		err = errors.New("can't find method " + methodName)
	}
	return
}

// Close 关闭服务器
func (server *Server) Close() error {
	return server.transport.Close()
}
