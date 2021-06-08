package controller

import (
	"PHONE/common"
	"PHONE/server/global_server"
	"PHONE/server/model/Person"
	"database/sql"
	"errors"
	"net"
	"strings"
)

//SendStr 封装信息到客户端
func SendStr(conn net.Conn, message string) error {
	mess, _ := common.Encode(message)
	_, err := conn.Write(mess)
	if err != nil {
		return err
	}
	return nil
}

func checkOnline(name string) (bool, string) {
	//global_server.Mutex.Lock()
	for _, client := range global_server.OnlineMap {
		if client.Name == name {
			return true, client.Ip
		}
	}
	//global_server.Mutex.Unlock()
	return false, ""
}

func removeAccount(pkg string) (option string, name string, password string) {
	list := strings.Split(pkg, common.Sep)
	option, name, password = list[0], list[1], list[2]
	return
}

func HandleAccount(conn net.Conn, DB *sql.DB, pkg string, client *global_server.Client, MMS *global_server.Message) bool {
	option, name, password := removeAccount(pkg)
	switch option {
	case "1":
		account, err := Person.Find(DB, name)
		if err != nil || account.Password != password {
			return false
		}
	case "2":
		person := Person.Person{
			Name:     name,
			Password: password,
		}
		if Person.Insert(DB, person) != nil {
			return false
		}
	default:
		return false
	}
	ok, _ := checkOnline(name)
	if ok == true {
		return false
	}
	*MMS = global_server.Message{
		Str:     "",
		Pow:     global_server.PowAll,
		MyIp:    conn.RemoteAddr().String(),
		OtherIp: "",
	}
	*client = global_server.Client{
		Ip:   conn.RemoteAddr().String(),
		Name: name,
		C:    make(chan string),
	}
	return true
}

// 组装信息字符串
func messageStr(name string, str string) string {
	return name + ":" + str
}

func HandleMessage(client *global_server.Client, message global_server.Message, str string) {
	message.Str = messageStr(client.Name, str)
	global_server.Msg <- message
}

func HandleSwt(message *global_server.Message, otherName string) error {
	if otherName == common.MyAll {
		message.Pow = global_server.PowAll
	} else {
		ok, OtherIp := checkOnline(otherName)
		if ok != true {
			return errors.New("NoFind")
		} else {
			message.Pow = global_server.PowOther
			message.OtherIp = OtherIp
		}
	}
	return nil
}

func showName(oldName, newName string) {
	message := global_server.Message{
		Str:     sysMessage(oldName + "已更名为:" + newName),
		Pow:     global_server.PowAll,
		MyIp:    "",
		OtherIp: "",
	}
	global_server.Msg <- message
}

func HandleName(DB *sql.DB, client *global_server.Client, newName string) error {
	err := Person.ReName(DB, client.Name, newName)
	if err != nil {
		return err
	}
	showName(client.Name, newName)
	client.Name = newName
	global_server.OnlineMap[client.Ip] = *client
	return nil
}

func HandlePassword(DB *sql.DB, client *global_server.Client, newPassword string) error {
	err := Person.RePassword(DB, client.Name, newPassword)
	if err != nil {
		return err
	}
	return nil
}

//组装发送文件字符串
func fileStr(name string, fileStr string) string {
	return name + common.Sep + fileStr
}

func HandleFile(client *global_server.Client, message global_server.Message, str string) {
	message.Str = fileStr(client.Name, str)
	message.IsFile = true
	global_server.Msg <- message
}

// sysMessage 组装系统消息
func sysMessage(str string) string {
	return "\t\t\t\t[系统消息]" + str
}

func NetWork() {
	for {
		message := <-global_server.Msg
		switch message.Pow {
		case global_server.PowAll:
			//global_server.Mutex.Lock()
			for _, v := range global_server.OnlineMap {
				v.C <- common.MyMessage + common.Sep + message.Str
			}
			//global_server.Mutex.Unlock()
		case global_server.PowMe:
			global_server.OnlineMap[message.MyIp].C <- common.MyMessage + common.Sep + message.Str
		case global_server.PowOther:
			if message.IsFile == true {
				global_server.OnlineMap[message.OtherIp].C <- common.MyFile + common.Sep + message.Str
			} else {
				global_server.OnlineMap[message.OtherIp].C <- common.MyMessage + common.Sep + "[悄悄话]" + message.Str
			}
		}
	}
}
