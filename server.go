package main

import (
	"chatroom/impl"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe(":7777", nil); err != nil {
		log.Println(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *impl.Connection
		str    string
		room   *impl.Room
		msg    impl.Message
	)
	//Upgrade:websocket  这里返回升级websocket确认
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}

	if conn, err = impl.InitConnection(wsConn); err != nil {
		conn.Close()
	}
	v, ok := impl.RoomMap.Load(conn.RoomId)
	if !ok { //房间不存在则初始化房间并存入RoomMap中
		room = impl.InitRoom()
		impl.RoomMap.Store(conn.RoomId, room)
	} else {
		room = v.(*impl.Room)
	}
	//将conn存入房间中
	room.PutConn(conn)
	str = fmt.Sprintf("%d号房间加入一名新用户", conn.RoomId)
	msg = impl.Message{ConnId: 0, Msg: str}
	room.PutMsg(msg)
}
