package ziface

type IMessage interface {
	GetDataLen() uint32
	GetMsgId() uint32
	GetData() []byte
	SetMsgId(uint32)
	SetData([]byte)    //设置消息内容
	SetDataLen(uint32) //设置消息数据端长度
}
