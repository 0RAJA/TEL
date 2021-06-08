package global_server

//权限
const (
	PowAll = iota
	PowMe
	PowOther
)

//上下线通知
const (
	In  = "已经上线"
	Out = "已经下线"
)

// Client 每个连接的信息
type Client struct {
	Ip   string
	Name string
	C    chan string
}

// Message 信息
type Message struct {
	Str     string //信息
	Pow     int    //权限
	MyIp    string //我的IP
	OtherIp string //目标IP
	IsFile  bool   //是否是文件
}
