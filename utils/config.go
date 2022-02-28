package utils

type GinxConf struct {
	Host             string //当前服务器主机IP
	TcpPort          int    //当前服务器主机监听端口号
	Name             string //当前服务器名称
	Version          string //当前Zinx版本号
	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 //业务工作Worker池的数量
	MaxWorkerTaskLen uint32 //业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    int
}
