package protocol

import (
	"encoding/binary"
	"errors"
)

const (
	MagicNumber uint32 = 0x5C2F3E1D // 魔数
	Version     byte   = 0x01       // 版本号
)

type MessageType byte

const (
	Request  MessageType = iota // 0
	Response                    // 1
)

// Header RPC消息头部
type Header struct {
	MagicNumber   uint32      // 固定标识（如 0x5RPC）
	Version       byte        // 协议版本（1）
	MessageType   MessageType // 消息类型（请求/响应）
	SerializeType byte        // 0=JSON, 1=Protobuf
	ServiceLength uint16      // 服务名长度
	MethodLength  uint16      // 方法名长度
	PayloadLength uint32      // 参数数据长度
}

// HeaderSize 消息头大小常量
const HeaderSize = 15 // 4+1+1+1+2+2+4 = 15 bytes

// EncodeHeader 将消息头编码为字节数组
func EncodeHeader(h *Header) []byte {
	buffer := make([]byte, HeaderSize) // 4+1+1+1+2+2+4 = 15 bytes

	binary.BigEndian.PutUint32(buffer[0:4], h.MagicNumber)
	buffer[4] = h.Version
	buffer[5] = byte(h.MessageType)
	buffer[6] = h.SerializeType
	binary.BigEndian.PutUint16(buffer[7:9], h.ServiceLength)
	binary.BigEndian.PutUint16(buffer[9:11], h.MethodLength)
	binary.BigEndian.PutUint32(buffer[11:15], h.PayloadLength)

	return buffer
}

// DecodeHeader 从字节数组解码消息头
func DecodeHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("invalid header data: too short")
	}

	h := &Header{
		MagicNumber:   binary.BigEndian.Uint32(data[0:4]),
		Version:       data[4],
		MessageType:   MessageType(data[5]),
		SerializeType: data[6],
		ServiceLength: binary.BigEndian.Uint16(data[7:9]),
		MethodLength:  binary.BigEndian.Uint16(data[9:11]),
		PayloadLength: binary.BigEndian.Uint32(data[11:15]),
	}

	if h.MagicNumber != MagicNumber {
		return nil, errors.New("invalid magic number")
	}

	return h, nil
}

// RequestMessage 请求消息
type RequestMessage struct {
	ServiceMethod string      // 格式: "Service.Method"
	Args          interface{} // 参数
}

// ResponseMessage 响应消息
type ResponseMessage struct {
	Error  string      // 错误信息，如果调用成功则为空
	Result interface{} // 结果
}
