package global_client

const HelpMessage = `退出: 	\q
修改姓名: \rN
修改密码: \rP
发送文件: \file
切换聊天模式: 输入@ XXX 或者 @ @all 对应公聊和私聊模式:
普通消息: 直接输入
"/\\,+=->*%!@^()|"`

//FILEPATH 文件默认路径
const FILEPATH = "."

// NotLegalString 不合法姓名或者密码
const NotLegalString = "/\\,`+=->*%!@^()|"

//接受返回验证值的通道
var (
	AccountOK  = make(chan bool)
	NameOK     = make(chan bool)
	PasswordOK = make(chan bool)
	SwtOK      = make(chan bool)
)

// FileFlag 确定是不是私聊模式,保证文件给一个人
var FileFlag = false
