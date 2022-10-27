package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/better-maksim/ginx/ziface"
)

type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer //当前conn属于哪个server，在conn初始化的时候添加即可
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前链接的id，也可以称作为 session id，id全局唯一
	ConnID uint32
	//当前连接的关闭状态
	isClosed bool
	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte
	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, server.GetGinxConf().MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

/*
 写消息Goroutine，用户讲数据发送给客户端
*/

func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//针对有缓冲channel需要些的数据处理
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				break
				fmt.Println("msgBuffChan is Closed")
			}
		case <-c.ExitBuffChan:
			//conn已经关闭
			return
		}
	}
}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is runing")
	defer fmt.Println(c.Conn.RemoteAddr().String(), "connreader exit!")
	defer c.Stop()

	for {
		//创建拆包解包的对象
		dp := NewDataPack(c.TcpServer.GetGinxConf().MaxPacketSize)
		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())

		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitBuffChan <- true
			continue
		}
		//拆包
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}
		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)
		//得到当前客户端请求的request 数据
		req := Reqeust{
			conn: c,
			msg:  msg,
		}

		if c.TcpServer.GetGinxConf().WorkerPoolSize > 0 {
			//已经启动了工作池的话，讲消息交给worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从路由Routers 中找到注册绑定Conn的对应Handle
			go c.MsgHandler.DoMsgHandler(&req)
		}

	}
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	dp := NewDataPack(c.TcpServer.GetGinxConf().MaxPacketSize)
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgBuffChan <- msg
	return nil
}

//Start 启动链接
func (c *Connection) Start() {
	//开启处理该链接督导客户端的数据后的请求业务
	go c.StartReader()
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			//得到消息不再阻塞
			return
		}
	}

}

func (c *Connection) Stop() {
	if c.isClosed == true {
		return

	}
	c.isClosed = true
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)
	// 关闭socket链接
	c.Conn.Close()
	//关闭Writer Goroutine
	c.ExitBuffChan <- true
	//将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c) //删除conn从ConnManager中
	//关闭该链接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//SendMsg 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	//1. 判断链接是否关闭，如果已经关闭了，则不应该继续发送
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	//将data封包，并且发送
	dp := NewDataPack(c.TcpServer.GetGinxConf().MaxPacketSize)
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}
	//写回客户端
	c.msgChan <- msg
	return nil
}

//SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

//GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}
