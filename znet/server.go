package znet

import (
	"fmt"
	"net"
	"time"

	"github.com/better-maksim/ginx/utils"
	"github.com/better-maksim/ginx/ziface"
)

type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler ziface.IMsgHandle
	//当前所有的链接
	ConnMgr ziface.IConnManager
	// =======================
	//新增两个hook函数原型
	//该Server的连接创建时Hook函数
	OnConnStart func(conn ziface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn ziface.IConnection)

	Config *utils.GinxConf
}

func (s *Server) GetGinxConf() *utils.GinxConf {
	return s.Config
}
func NewServer(conf *utils.GinxConf) ziface.IServer {
	s := &Server{
		Config:     conf,
		Name:       conf.Name,
		IPVersion:  "tcp4",
		IP:         conf.Host,
		Port:       conf.TcpPort,
		msgHandler: NewMsgHandle(conf.WorkerPoolSize, conf.MaxMsgChanLen),
		ConnMgr:    NewConnManager(), //创建ConnManager
	}
	return s
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}

// Start 启动服务
func (s *Server) Start() {
	// 如果我们的方法正确，运行时就会打印出Version: V0.4, MaxConn: 3,  MaxPacketSize: 4096
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		s.Config.Version,
		s.Config.MaxConn,
		s.Config.MaxPacketSize)

	//开启第一个go 去做服务端的lister业务
	go func() {
		//0. 启动worker pool
		s.msgHandler.StartWorkerPool()

		//1. 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))

		if err != nil {
			fmt.Println("listten", s.IPVersion, "err", err)
			return
		}

		//2. 监听listenner
		listenner, err := net.ListenTCP(s.IPVersion, addr)

		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}
		//已经监听成功
		fmt.Println("start "+s.Name+" server  ", s.Name, " succ, now listenning...")

		//TODO server.go 应该有一个自动生成id的方法
		var cid uint32
		cid = 0
		//3. 启动 server 网络链接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//3.2 设置服务器最大链接控制，如果超过最大链接，那么关闭此新的链接
			if s.ConnMgr.Len() >= s.Config.MaxConn {
				conn.Close()
				continue
			}
			//3.3 Server.sStart() 处理新的链接请求业务方法，此时应该有hanndler和conn是绑定的
			//我们这里暂时做一个最大512字节的回显服务
			dealConn := NewConntion(s, conn, cid, s.msgHandler)
			cid++
			go dealConn.Start()
		}

	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)
	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()
	//TODO： Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加
	//阻塞,否则主Go退出， listenner的go将会退出
	for {
		time.Sleep(10 * time.Second)
	}
}

//路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}
