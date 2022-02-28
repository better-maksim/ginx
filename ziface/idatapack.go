package ziface

/**
封装数据和拆包数据
直接面向TCP链接中的数据流，为传输数据添加头部信息，用于处理TCP粘包问题。
*/

type IDataPack interface {
	GetHeadLen() uint32
	Pack(msg IMessage) ([]byte, error)
	Unpack([]byte) (IMessage, error)
}
