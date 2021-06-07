package main

import (
	"PHONE/common"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial(common.NETWORK, common.ADDRESS)
	if err != nil {
		fmt.Println("连接失败,请稍后再次尝试")
	}
	defer conn.Close()

}


