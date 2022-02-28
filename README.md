# Ginx

该库来源于 `zinx`：

> https://github.com/aceld/zinx

在哔哩哔哩上有视频讲解，内容和能详细，大家可以去看一下，文字版可以到实验楼进行学习，这些都是免费的。

为了更符合我个人的开发习惯，将会增加很多自己用到的特性，所以并没有给 `zinx` 提交代码，而是单独开了一个新的库，大家可以去给 `zinx` 点个 Star。

# 快速起步

## echo server

```bash
go get github.com/better-maksim/ginx
```

server.go

```
// EchoRouter ping test 自定义路由
type EchoRouter struct {
	znet.BaseRouter //一定要先基础BaseRouter
}

// Handle Test Handle
func (this *EchoRouter) Handle(request ziface.IRequest) {
	echoData := string(request.GetData())
	err := request.GetConnection().SendBuffMsg(1, []byte(echoData))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

func main() {
	s := znet.NewServer(&utils.GinxConf{
		Name:             "echo server",
		Version:          "V0.11",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          12000,
		MaxPacketSize:    4096,
		WorkerPoolSize:   4,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
	})
	//配置路由
	s.AddRouter(0, &EchoRouter{})
	//开启服务
	s.Serve()
}
```


client.go

```
package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/better-maksim/ginx/znet"
)

/*
   模拟客户端
*/
func main() {
	data := []byte("Echo Client Test Message")

	fmt.Println("Client Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:8999")

	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}
	for {
		//发封包message消息
		dp := znet.NewDataPack(0)
		msg, _ := dp.Pack(znet.NewMsgPackage(0, data))

		_, err := conn.Write(msg)

		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())
			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}
			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		time.Sleep(1 * time.Second)
	}
}
```