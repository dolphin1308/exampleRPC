package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rpc/codec"
	"rpc/example"
	"rpc/server"
	"rpc/transport"
)

var (
	addr           = flag.String("addr", ":8972", "服务地址")
	transportType  = flag.String("transport", "tcp", "传输协议 (tcp/http)")
	serializerType = flag.String("serializer", "json", "序列化协议 (json/protobuf)")
)

func main() {
	flag.Parse()

	// 解析传输类型
	var tType transport.TransportType
	switch *transportType {
	case "tcp":
		tType = transport.TCP
		fmt.Println("使用TCP传输协议")
	case "http":
		tType = transport.HTTP
		fmt.Println("使用HTTP传输协议")
	default:
		log.Fatalf("不支持的传输协议: %s", *transportType)
	}

	// 解析序列化类型
	var cType codec.Type
	switch *serializerType {
	case "json":
		cType = codec.JSON
		fmt.Println("使用JSON序列化")
	case "protobuf":
		cType = codec.Protobuf
		fmt.Println("使用Protobuf序列化")
	default:
		log.Fatalf("不支持的序列化协议: %s", *serializerType)
	}

	// 创建RPC服务器
	s := server.NewServer(tType, cType)

	// 注册服务
	err := s.Register(new(example.ArithService))
	if err != nil {
		log.Fatal("注册算术服务失败:", err)
	}
	fmt.Println("已注册算术服务")

	err = s.Register(new(example.EchoService))
	if err != nil {
		log.Fatal("注册Echo服务失败:", err)
	}
	fmt.Println("已注册Echo服务")

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("正在关闭服务器...")
		s.Close()
		os.Exit(0)
	}()

	// 启动服务器
	fmt.Printf("RPC服务器正在监听 %s\n", *addr)
	if err := s.Serve(*addr); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
