package codec

import (
	"errors"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// ProtobufCodec 实现了基于Protobuf的编解码器
type ProtobufCodec struct{}

// Encode 将Protobuf消息编码为字节数组
func (p *ProtobufCodec) Encode(value interface{}) ([]byte, error) {
	// 检查是否为proto.Message类型
	message, ok := value.(proto.Message)
	if !ok {
		return nil, errors.New("value is not a proto.Message")
	}
	return proto.Marshal(message)
}

// Decode 将字节数组解码为Protobuf消息
func (p *ProtobufCodec) Decode(data []byte, value interface{}) error {
	// 检查是否为proto.Message类型
	message, ok := reflect.ValueOf(value).Elem().Interface().(proto.Message)
	if !ok {
		return errors.New("value is not a proto.Message")
	}
	// 先获取指向message的指针，然后使用proto.Unmarshal
	return proto.Unmarshal(data, message)
}
