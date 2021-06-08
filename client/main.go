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
	conn, err := net.Dial(common.NETWORK, common.ADDRESS)
	if err != nil {
		fmt.Println("连接失败,请稍后再次尝试")
		return
	}
	defer conn.Close()
	go controller.Handle(conn) //接受消息
	if controller.TestAccount(conn) == false {
		return
	}
	defer Quit(conn)
	fmt.Println("\t\t\t\t回车发送消息,\\help获取帮助")
	inputReader := bufio.NewReader(os.Stdin)
	for {
		inputReader.Reset(os.Stdin)
		input := enterTheStr(inputReader)
		switch input {
		case common.MyQuit:
			return
		case common.MyReName:
			controller.SendName(conn)
		case common.MyRePassword:
			controller.SendPassword(conn)
		case common.MyHelp:
			fmt.Println(global.HelpMessage)
		case common.MySwt:
			controller.SendSwt(conn)
		case common.MyFile:
			if global.FileFlag == false {
				fmt.Println("您现在处于群聊模式,无法发送文件")
			} else {
				controller.SendFile(conn)
			}
		default:
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

func Quit(conn net.Conn) {
	_ = controller.SendStr(conn, common.MyQuit+common.Sep+common.OK)
	<-global.QuitOK
	fmt.Println("再见")
}
