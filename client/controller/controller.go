package controller

import (
	"PHONE/client/global_client"
	"PHONE/common"
	"bufio"
	"fmt"
	"net"
	"os"
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

func removeMessage(pkg string) (cmd, message string) {
	list := strings.Split(pkg, common.Sep)
	cmd = list[0]
	message = strings.Join(list[1:], common.Sep)
	return
}

func AtoB(str string) bool {
	if str == common.OK {
		return true
	}
	return false
}

func Handle(conn net.Conn) {
	handMessage := func(pkg string) {
		cmd, message := removeMessage(pkg)
		switch cmd {
		case common.TestAccount:
			global_client.AccountOK <- AtoB(message)
		case common.MyMessage:
			fmt.Println(message)
		case common.MySwt:
			global_client.SwtOK <- AtoB(message)
		case common.MyReName:
			global_client.NameOK <- AtoB(message)
		case common.MyRePassword:
			global_client.PasswordOK <- AtoB(message)
		case common.MyFile:
			go ReceiveFile(message)
		}
	}
	for {
		readBuf := bufio.NewReader(conn)
		for {
			message, err := common.Decode(readBuf)
			if err != nil {
				break
			}
			handMessage(message)
		}
	}
}

func IsLegal(str string) bool {
	if strings.ContainsAny(str, global_client.NotLegalString) || len(str) > 15 {
		return false
	}
	return true
}

func PkgAccount(option, name, password string) string {
	return common.TestAccount + common.Sep + option + common.Sep + name + common.Sep + password
}

func TestAccount(conn net.Conn) bool {
	var (
		name     string
		password string
	)
	for {
		fmt.Println("1.登录账号\n2.注册账号\n3.离开")
		var option string
		_, err := fmt.Scan(&option)
		if err != nil {
			fmt.Println("输入有误,请重新输入")
			continue
		}
		switch option {
		case "1", "2":
			for {
				fmt.Println("输入用户名(" + common.MyQuit + "退出):")
				_, err := fmt.Scan(&name)
				if err != nil {
					fmt.Println("输入有误,请重新输入")
					continue
				}
				if name == common.MyQuit {
					break
				}
				fmt.Println("输入密码")
				_, err = fmt.Scan(&password)
				if err != nil {
					fmt.Println("输入有误,请重新输入")
					continue
				}
				if IsLegal(name) && IsLegal(password) {
					err := SendStr(conn, PkgAccount(option, name, password))
					if err != nil {
						fmt.Println("登录或注册失败,请重新尝试")
						continue
					}
					ok := <-global_client.AccountOK
					if ok != true {
						fmt.Println("登录或注册失败,请重新输入")
					} else {
						fmt.Println("欢迎" + name + "回来")
						return true
					}
				} else {
					fmt.Println("姓名或者密码格式有误,请重新输入")
					continue
				}
			}
		case "3":
			return false
		default:
			fmt.Println("输入有误,请重新输入")
		}
	}
}

func SendMessage(conn net.Conn, message string) {
	_ = SendStr(conn, common.MyMessage+common.Sep+message)
}

func SendName(conn net.Conn) {
	var newName string
	fmt.Println("输入新的姓名")
	_, err := fmt.Scan(&newName)
	if err != nil || IsLegal(newName) != true {
		fmt.Println("姓名输入有误")
		return
	}
	_ = SendStr(conn, common.MyReName+common.Sep+newName)
	ok := <-global_client.NameOK
	if ok != true {
		fmt.Println("修改失败,姓名重复")
	} else {
		fmt.Println("修改成功")
	}
}

func SendPassword(conn net.Conn) {
	var newPassword string
	fmt.Println("输入新的密码")
	_, err := fmt.Scan(newPassword)
	if err != nil || IsLegal(newPassword) {
		fmt.Println("输入有误,请核对后输入")
		return
	}
	_ = SendStr(conn, common.MyRePassword+common.Sep+newPassword)
	ok := <-global_client.PasswordOK
	if ok != true {
		fmt.Println("修改失败")
	} else {
		fmt.Println("修改成功")
	}
}

func SendSwt(conn net.Conn) {
	var otherName string
	fmt.Println("输入需要切换的对象")
	_, err := fmt.Scan(&otherName)
	if err != nil {
		fmt.Println("输入有误,请重新输入")
		return
	}
	_ = SendStr(conn, common.MySwt+common.Sep+otherName)
	ok := <-global_client.SwtOK
	if ok != true {
		fmt.Println("切换失败")
		return
	}
	fmt.Println("切换成功")
	if otherName == common.MyAll {
		global_client.FileFlag = false
	} else {
		global_client.FileFlag = true
	}
}

func SendFile(conn net.Conn) {
	var path string
	fmt.Println("输入文件路径")
	_, err := fmt.Scan(&path)
	if err != nil {
		fmt.Println("输入有误")
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		fmt.Println("无法打开文件")
		return
	}
	if info.Size() > common.SendMaxSize {
		fmt.Printf("文件超过 %d MB\n", common.SendMaxSize/(1024*1024))
		return
	}
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("文件打开失败")
		return
	}
	defer file.Close()
	fileStr := make([]byte, common.SendMaxSize)
	n, err := file.Read(fileStr)
	err = SendStr(conn, common.MyFile+common.Sep+info.Name()+common.Sep+string(fileStr[:n]))
	if err != nil {
		fmt.Println("发送失败")
		return
	}
	fmt.Println("发送成功")
}

func remove(pkg string) (name, fileName, fileStr string) {
	list := strings.Split(pkg, common.Sep)
	name, fileName, fileStr = list[0], list[1], strings.Join(list[2:], common.Sep)
	return
}

func ReceiveFile(pkg string) {
	name, fileName, fileStr := remove(pkg)
	file, err := os.Create(global_client.FILEPATH + "\\" + fileName)
	if err != nil {
		fmt.Println("您有一个来自", name, "的文件接受失败")
		return
	}
	defer file.Close()
	_, err = file.Write([]byte(fileStr))
	if err != nil {
		fmt.Println("您有一个来自", name, "的文件保存失败")
		return
	}
	fmt.Println("您有一个来自", name, "的文件")
	return
}
