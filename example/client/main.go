package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"rpc/client"
	"rpc/codec"
	"rpc/example"
	"rpc/transport"
)

var (
	serverAddr     = flag.String("addr", "localhost:8972", "服务器地址")
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

	// 配置客户端
	opt := &client.Option{
		TransportType: tType,
		CodecType:     cType,
		Timeout:       time.Second * 5,
	}

	// 创建客户端
	c := client.NewClient(*serverAddr, opt)
	defer c.Close()

	fmt.Printf("连接到RPC服务器 %s\n", *serverAddr)

	// 测试算术服务
	testArithService(c)

	// 测试Echo服务
	testEchoService(c)
}

// testArithService 测试算术服务
func testArithService(c *client.Client) {
	// 测试加法
	args := example.Args{A: 10, B: 20}
	var reply example.Result

	err := c.Call("ArithService.Add", args, &reply)
	if err != nil {
		log.Fatalf("调用ArithService.Add错误: %v", err)
	}
	fmt.Printf("ArithService.Add: %d + %d = %d\n", args.A, args.B, reply.Value)

	// 测试减法
	err = c.Call("ArithService.Sub", args, &reply)
	if err != nil {
		log.Fatalf("调用ArithService.Sub错误: %v", err)
	}
	fmt.Printf("ArithService.Sub: %d - %d = %d\n", args.A, args.B, reply.Value)

	// 测试乘法
	err = c.Call("ArithService.Mul", args, &reply)
	if err != nil {
		log.Fatalf("调用ArithService.Mul错误: %v", err)
	}
	fmt.Printf("ArithService.Mul: %d * %d = %d\n", args.A, args.B, reply.Value)

	// 测试除法
	err = c.Call("ArithService.Div", args, &reply)
	if err != nil {
		log.Fatalf("调用ArithService.Div错误: %v", err)
	}
	fmt.Printf("ArithService.Div: %d / %d = %d\n", args.A, args.B, reply.Value)

	// 测试除零错误
	divZeroArgs := example.Args{A: 10, B: 0}
	err = c.Call("ArithService.Div", divZeroArgs, &reply)
	if err != nil {
		fmt.Printf("期望的除零错误: %v\n", err)
	} else {
		log.Fatalf("除零操作未返回错误")
	}
}

// testEchoService 测试Echo服务
func testEchoService(c *client.Client) {
	args := example.EchoArgs{Message: "Hello, RPC!"}
	var reply example.EchoResult

	err := c.Call("EchoService.Echo", args, &reply)
	if err != nil {
		log.Fatalf("调用EchoService.Echo错误: %v", err)
	}
	fmt.Printf("EchoService.Echo: 发送 '%s', 接收 '%s'\n", args.Message, reply.Message)
}
