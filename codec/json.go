package codec

import "encoding/json"

// JSONCodec 实现了基于JSON的编解码器
type JSONCodec struct{}

// Encode 将对象编码为JSON字节数组
func (j *JSONCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// Decode 将JSON字节数组解码为对象
func (j *JSONCodec) Decode(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
