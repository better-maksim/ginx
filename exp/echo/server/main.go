package main

import (
	"fmt"
	"github.com/better-maksim/ginx/utils"
	"github.com/better-maksim/ginx/ziface"
	"github.com/better-maksim/ginx/znet"
)

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
