package example

import (
	"errors"
	"fmt"
)

// Args 计算服务参数
type Args struct {
	A int
	B int
}

// Result 计算结果
type Result struct {
	Value int
}

// ArithService 算数服务
type ArithService struct{}

// Add 加法操作
func (a *ArithService) Add(args Args, result *Result) error {
	result.Value = args.A + args.B
	return nil
}

// Sub 减法操作
func (a *ArithService) Sub(args Args, result *Result) error {
	result.Value = args.A - args.B
	return nil
}

// Mul 乘法操作
func (a *ArithService) Mul(args Args, result *Result) error {
	result.Value = args.A * args.B
	return nil
}

// Div 除法操作
func (a *ArithService) Div(args Args, result *Result) error {
	if args.B == 0 {
		return errors.New("division by zero")
	}
	result.Value = args.A / args.B
	return nil
}

// Echo 字符串响应服务
type EchoService struct{}

// EchoArgs 请求参数
type EchoArgs struct {
	Message string
}

// EchoResult 响应结果
type EchoResult struct {
	Message string
}

// Echo 返回输入的字符串
func (e *EchoService) Echo(args EchoArgs, result *EchoResult) error {
	result.Message = fmt.Sprintf("Echo: %s", args.Message)
	return nil
}
