package codec

// Codec 定义序列化和反序列化的接口
type Codec interface {
	Encode(value interface{}) ([]byte, error)    // 编码数据
	Decode(data []byte, value interface{}) error // 解码数据
}

// Type 表示序列化类型
type Type byte

const (
	JSON     Type = iota // 0
	Protobuf             // 1
)

// NewCodec 根据编解码类型创建对应的编解码器
func NewCodec(codecType Type) Codec {
	switch codecType {
	case JSON:
		return &JSONCodec{}
	case Protobuf:
		return &ProtobufCodec{}
	default:
		return &JSONCodec{} // 默认使用JSON
	}
}
