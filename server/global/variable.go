package global

import "sync"

var OnlineMap = make(map[string]Client) //广播名单
var Msg = make(chan Message)            //负责给广播站信息
var Mutex sync.Mutex
