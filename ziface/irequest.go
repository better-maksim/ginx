package ziface

//I Request接口
// 实际上把客户端请求的连接信息和请求数据包装到了request里
type IRequest interface {
	GetConnection() IConnection
	GetData() []byte
	GetMsgId() uint32
}
