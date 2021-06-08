package main

import (
	"PHONE/client/controller"
	global "PHONE/client/global_client"
	"PHONE/common"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial(common.NETWORK, common.ADDRESS) //拨号
	if err != nil {
		fmt.Println("连接失败,请稍后再次尝试")
		return
	}
	defer conn.Close()
	fmt.Println(global.HelpMessage)
	go controller.Handle(conn)                 //接受消息
	if controller.TestAccount(conn) == false { //验证账户
		return
	}
	defer Quit(conn) //程序结束时要发送退出请求
	fmt.Println("\t\t\t\t回车发送消息,\\help获取帮助")
	inputReader := bufio.NewReader(os.Stdin)
	for {
		inputReader.Reset(os.Stdin)       //重置缓冲区
		input := enterTheStr(inputReader) //输入数据
		switch input {
		case common.MyQuit: //退出
			return
		case common.MyReName: //改名
			controller.SendName(conn)
		case common.MyRePassword: //改密码
			controller.SendPassword(conn)
		case common.MyHelp: //帮助文档
			fmt.Println(global.HelpMessage)
		case common.MySwt: //切换聊天模式
			controller.SendSwt(conn)
		case common.MyFile: //传文件
			if global.FileFlag == false {
				fmt.Println("您现在处于群聊模式,无法发送文件")
			} else {
				controller.SendFile(conn)
			}
		default: // 传普通消息
			if len(input) == 0 || len(input) > common.SendMaxSize {
				fmt.Println("字数有误,请重新输入")
			} else {
				controller.SendMessage(conn, input)
			}
		}
	}
}

// enterTheStr 从键盘读取字符串
func enterTheStr(inputReader *bufio.Reader) string {
	buf, _ := inputReader.ReadString('\n')
	buf = strings.Trim(buf, "\r\n")
	return buf
}

// Quit 退出
func Quit(conn net.Conn) {
	_ = controller.SendStr(conn, common.MyQuit+common.Sep+common.OK)
	fmt.Println("再见")
}
