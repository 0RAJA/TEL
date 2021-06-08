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
	listen, err := net.Listen(NETWORK, ADDRESS)
	if err != nil {
		log.Println(err)
		return
	}
	defer listen.Close()
	DB, err := Person.DBInit()
	if err != nil {
		log.Println(err)
		return
	}
	go controller.NetWork()
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go process(conn, DB)
	}
}

func process(conn net.Conn, DB *sql.DB) {
	defer conn.Close()
	var (
		MMS    global_server.Message
		client global_server.Client
	)
	defer func() {
		delete(global_server.OnlineMap, client.Ip)
		InAndOut(client.Name, global_server.Out)
	}()
	for {
		readBuf := bufio.NewReaderSize(conn, SendMaxSize)
		for {
			pkg, err := Decode(readBuf)
			if err != nil {
				break
			} else {
				fmt.Println(pkg)
				if Handle(pkg, &client, &MMS, DB, conn) != nil {
					fmt.Println(client.Name)
					return
				}
			}
		}
	}
}

func removeCMD(pkg string) (cmd string, message string) {
	list := strings.Split(pkg, Sep)
	cmd = list[0]
	message = strings.Join(list[1:], Sep)
	return cmd, message
}

func Handle(pkg string, client *global_server.Client, MMS *global_server.Message, DB *sql.DB, conn net.Conn) error {
	cmd, message := removeCMD(pkg)
	switch cmd {
	case TestAccount:
		if controller.HandleAccount(conn, DB, message, client, MMS) != true {
			_ = reply(conn, TestAccount, NO)
		} else {
			_ = reply(conn, TestAccount, OK)
			global_server.OnlineMap[client.Ip] = *client
			go SynMessage(conn, client)
			InAndOut(client.Name, global_server.In)
		}
	case MyQuit:
		_ = reply(conn, MyQuit, OK)
		return errors.New(MyQuit)
	case MyMessage:
		controller.HandleMessage(client, *MMS, message)
	case MyReName:
		err := controller.HandleName(DB, client, message)
		if err != nil {
			err = reply(conn, MyReName, NO)
		} else {
			err = reply(conn, MyReName, OK)
		}
	case MyRePassword:
		err := controller.HandlePassword(DB, client, message)
		if err != nil {
			err = reply(conn, MyRePassword, NO)
		} else {
			err = reply(conn, MyRePassword, OK)
		}
	case MySwt:
		err := controller.HandleSwt(MMS, message)
		if err != nil {
			err = reply(conn, MySwt, NO)
		} else {
			err = reply(conn, MySwt, OK)
		}
	case MyFile:
		controller.HandleFile(client, *MMS, message)
	}
	return nil
}

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
