package main

import (
	"chatroom/impl"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	//思路：当一个用户连接时，为他分配到一个房间内，
	roomMap = make(map[int]impl.Room) //房间集合
)

func main() {
	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe(":7777", nil); err != nil {
		log.Println(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("hello"))
	var (
		wsConn *websocket.Conn
		err    error
		conn   *impl.Connection
		str    string
	)
	//Upgrade:websocket  这里返回升级websocket确认
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}

	if conn, err = impl.InitConnection(wsConn); err != nil {
		conn.Close()
	}

	//写数据
	go func() {
		for {
			str = fmt.Sprintf("服务器时间：%s", time.Now().Format("2006-01-02 15:04:05"))
			if err := conn.WriteMessage([]byte(str)); err != nil {
				log.Println(err)
				goto ERR
			}
			log.Println("服务端：", str)
			time.Sleep(time.Second * 5)
		}
	ERR:
		fmt.Println("wsHandler写数据")
		conn.Close()
	}()

	//读数据
	go func() {
		var (
			data []byte
			err  error
		)
		for {
			if data, err = conn.ReadMessage(); err != nil {
				log.Println(err)
				goto ERR
			}
			log.Println("客户端：", string(data))
		}
	ERR:
		fmt.Println("wsHandler读数据")
		conn.Close()
	}()

}

func wsHandler2(w http.ResponseWriter, r *http.Request) {
	var (
		conn *websocket.Conn
		err  error
		data []byte
	)
	conn, err = upgrader.Upgrade(w, r, nil)
	for {
		log.Println("for:")
		//Text,Binary
		if _, data, err = conn.ReadMessage(); err != nil {
			//关闭连接
			goto ERR
		}
		log.Println("data:", data)
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
