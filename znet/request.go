package znet

import (
	"github.com/better-maksim/ginx/ziface"
)

type Reqeust struct {
	conn ziface.IConnection
	msg  ziface.IMessage
}

func (r *Reqeust) GetConnection() ziface.IConnection {
	return r.conn
}

func (r *Reqeust) GetData() []byte {
	return r.msg.GetData()
}

func (r *Reqeust) GetMsgId() uint32 {
	return r.msg.GetMsgId()
}
