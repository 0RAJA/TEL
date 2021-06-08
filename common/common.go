package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

//监听口
const (
	NETWORK = "tcp"
	ADDRESS = "127.0.0.1:8800"
)

//命令
const (
	MyQuit       = "\\q"
	MyReName     = "\\rN"
	MyRePassword = "\\rP"
	MyFile       = "\\file"

	MySwt     = "@"
	MyAll     = "@all"
	MyMessage = "\\message"
	MyHelp    = "\\help"

	TestAccount = "\\Account"
)

//接受消息的返回验证值
const (
	OK = "true"
	NO = "false"
)

// SendMaxSize 文件最大字节数
const (
	SendMaxSize    = 10495760 //最大发送
)

const Sep = "->"

// Encode 将消息编码
func Encode(message string) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）1字节8位
	var length = int32(len(message))
	var pkg = new(bytes.Buffer) //创建一个缓冲区
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	err = binary.Write(pkg, binary.LittleEndian, []byte(message))
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// Decode 解码消息
func Decode(reader *bufio.Reader) (string, error) {
	// 读取消息的长度
	lengthByte, _ := reader.Peek(4) // 读取前4个字节的数据
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length) //读这个缓冲区并用规定方式解码,返回解码结果给length(其实就是返回前四个字节代表的数字)
	if err != nil {
		return "", err
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 { //判断对不对
		return "", err
	}
	// 读取真正的消息数据
	pack := make([]byte, int(4+length)) //就建这么大,读这么多
	_, err = reader.Read(pack)
	if err != nil {
		return "", err
	}
	return string(pack[4:]), nil //拆包返回
}

