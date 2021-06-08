package main

import (
	. "PHONE/common"
	"PHONE/server/controller"
	"PHONE/server/global_server"
	"PHONE/server/model/Person"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	listen, err := net.Listen(NETWORK, ADDRESS) //监听IP端口
	if err != nil {
		log.Println(err)
		return
	}
	defer listen.Close()
	DB, err := Person.DBInit() //连接数据库
	if err != nil {
		log.Println(err)
		return
	}
	go controller.NetWork() //启动接受信息平台
	for {
		conn, err := listen.Accept() //建立连接
		if err != nil {
			continue
		}
		go process(conn, DB) //处理请求
	}
}

func process(conn net.Conn, DB *sql.DB) {
	defer conn.Close()
	var (
		MMS    global_server.Message //消息模板
		client global_server.Client  //连接信息
	)
	defer func() { //此用户退出时更新通知列表,发送退出消息,删除通知map,关闭C channel
		delete(global_server.OnlineMap, client.Ip)
		InAndOut(client.Name, global_server.Out)
		close(client.C)
	}()
	for {
		readBuf := bufio.NewReaderSize(conn, SendMaxSize) //建立缓冲区
		for {
			pkg, err := Decode(readBuf) //对信息拆包
			if err != nil {
				break
			} else {
				fmt.Println(pkg)
				if Handle(pkg, &client, &MMS, DB, conn) != nil { //判断连接是否断开
					return
				}
			}
		}
	}
}

//对信息进行拆包
func removeCMD(pkg string) (cmd string, message string) {
	list := strings.Split(pkg, Sep)
	cmd = list[0]
	message = strings.Join(list[1:], Sep)
	return cmd, message
}

// Handle 对各个连接的信息进行处理
func Handle(pkg string, client *global_server.Client, MMS *global_server.Message, DB *sql.DB, conn net.Conn) error {
	cmd, message := removeCMD(pkg) //命令+内容
	switch cmd {
	case TestAccount: //验证账户
		if controller.HandleAccount(conn, DB, message, client, MMS) != true {
			_ = reply(conn, TestAccount, NO) //验证失败回复NO
		} else {
			_ = reply(conn, TestAccount, OK)             //验证成功回复OK
			global_server.OnlineMap[client.Ip] = *client //加入广播map
			go SynMessage(conn, client)                  //启动信息同步平台
			InAndOut(client.Name, global_server.In)
		}
	case MyQuit: //退出
		return errors.New(MyQuit)
	case MyMessage: //信息
		controller.HandleMessage(client, *MMS, message)
	case MyReName: //改名指令
		err := controller.HandleName(DB, client, message)
		if err != nil {
			err = reply(conn, MyReName, NO)
		} else {
			err = reply(conn, MyReName, OK)
		}
	case MyRePassword: //改密码指令
		err := controller.HandlePassword(DB, client, message)
		if err != nil {
			err = reply(conn, MyRePassword, NO)
		} else {
			err = reply(conn, MyRePassword, OK)
		}
	case MySwt: //切换聊天模式
		err := controller.HandleSwt(MMS, message)
		if err != nil {
			err = reply(conn, MySwt, NO)
		} else {
			err = reply(conn, MySwt, OK)
		}
	case MyFile: //传输文件
		controller.HandleFile(client, *MMS, message)
	}
	return nil
}

//回复客户端是否成功的验证
func reply(conn net.Conn, cmd string, result string) error {
	err := controller.SendStr(conn, cmd+Sep+result)
	if err != nil {
		return err
	}
	return nil
}

//SynMessage 监听广播站,传回信息
func SynMessage(conn net.Conn, clint *global_server.Client) {
	for message := range clint.C {
		err := controller.SendStr(conn, message)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

//登录退出字符串的组装
func inOutStr(name string, str string) string {
	t := time.Now()
	s := t.Format("2006年1月2日 15:04:05 ")
	return "\t\t\t\t[系统消息]" + s + name + str
}

//InAndOut 通知登录退出
func InAndOut(name string, str string) {
	message := global_server.Message{
		Str:     inOutStr(name, str),
		Pow:     global_server.PowAll,
		MyIp:    "",
		OtherIp: "",
	}
	global_server.Msg <- message
	OnlineNum()
}

//OnlineNum 通报当前在线人数
func OnlineNum() {
	message := global_server.Message{
		Str:     "\t\t\t\t[系统消息]" + "当前在线人数为:" + strconv.Itoa(len(global_server.OnlineMap)) + "人",
		Pow:     global_server.PowAll,
		OtherIp: "",
	}
	global_server.Msg <- message
}
