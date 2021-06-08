package controller

import (
	"PHONE/common"
	"PHONE/server/global_server"
	"PHONE/server/model/Person"
	"crypto/md5"
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

//检查是否在线
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

//对账户信息拆包
func removeAccount(pkg string) (option string, name string, password string) {
	list := strings.Split(pkg, common.Sep)
	option, name, password = list[0], list[1], list[2]
	return
}

// HandleAccount 验证账户
func HandleAccount(conn net.Conn, DB *sql.DB, pkg string, client *global_server.Client, MMS *global_server.Message) bool {
	option, name, password := removeAccount(pkg)
	switch option {
	case "1":
		account, err := Person.Find(DB, name)
		if err != nil {
			return false
		}
		hashName1 := Person.Md5Str(md5.Sum([]byte(password)))
		if hashName1 != account.Password {
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
	ok, _ := checkOnline(name) //检查是否有其他人登录
	if ok == true {
		return false
	}
	*MMS = global_server.Message{ //初始化信息模板
		Str:     "",
		Pow:     global_server.PowAll,
		MyIp:    conn.RemoteAddr().String(),
		OtherIp: "",
	}
	*client = global_server.Client{ //初始化连接信息
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

// HandleMessage 处理消息
func HandleMessage(client *global_server.Client, message global_server.Message, str string) {
	message.Str = messageStr(client.Name, str)
	global_server.Msg <- message
}

// HandleSwt 处理更改聊天模式
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

//更名通知
func showName(oldName, newName string) {
	message := global_server.Message{
		Str:     sysMessage(oldName + "已更名为:" + newName),
		Pow:     global_server.PowAll,
		MyIp:    "",
		OtherIp: "",
	}
	global_server.Msg <- message
}

// HandleName 处理改名
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

// HandlePassword 处理改密码
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

// HandleFile 处理文件
func HandleFile(client *global_server.Client, message global_server.Message, str string) {
	message.Str = fileStr(client.Name, str)
	message.IsFile = true
	global_server.Msg <- message
}

// sysMessage 组装系统消息
func sysMessage(str string) string {
	return "\t\t\t\t[系统消息]" + str
}

// NetWork 广播站,广播信息
func NetWork() {
	for {
		message := <-global_server.Msg
		switch message.Pow {
		case global_server.PowAll: //群发
			//global_server.Mutex.Lock()
			for _, v := range global_server.OnlineMap {
				v.C <- common.MyMessage + common.Sep + message.Str
			}
			//global_server.Mutex.Unlock()
		case global_server.PowMe: //尽自己可见
			global_server.OnlineMap[message.MyIp].C <- common.MyMessage + common.Sep + message.Str
		case global_server.PowOther: //私发
			if message.IsFile == true { //文件
				global_server.OnlineMap[message.OtherIp].C <- common.MyFile + common.Sep + message.Str
			} else {
				global_server.OnlineMap[message.OtherIp].C <- common.MyMessage + common.Sep + "[悄悄话]" + message.Str
			}
		}
	}
}
