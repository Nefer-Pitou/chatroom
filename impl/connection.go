package impl

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Connection struct {
	wsConn *websocket.Conn
	//inChan    chan Message
	outChan   chan Message
	closeChan chan byte
	IsClose   bool
	mutex     sync.Mutex
	RoomId    int //所在房间ID
	Id        int
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	roomId := time.Now().Second()%2 + 1
	id := time.Now().Nanosecond()
	conn = &Connection{
		wsConn: wsConn,
		//inChan:    make(chan Message, 1000),
		outChan:   make(chan Message, 1000),
		closeChan: make(chan byte, 1),
		RoomId:    roomId,
		Id:        id,
	}
	//启动读协程
	go conn.readLoop()
	//启动写协程
	go conn.writeLoop()

	return
}

func (conn *Connection) WriteMessage(msg Message) (err error) {
	select {
	case conn.outChan <- msg:
	case <-conn.closeChan:
		err = errors.New("WriteMessage():connection was disconnected")
	}
	return
}

func (conn *Connection) Close() {
	conn.mutex.Lock()
	if !conn.IsClose {
		close(conn.closeChan)
		conn.IsClose = true
		//线程安全，可重入的close
		if err := conn.wsConn.Close(); err != nil {
			log.Println("Connection.Close():", err)
		}
		//从RoomMap中的该conn所在房间中删掉该conn
		v, ok := RoomMap.Load(conn.RoomId)
		if !ok {
			log.Println("找不到", conn.RoomId, "号房间！")
		} else {
			//从房间中删除该conn
			v.(*Room).connMap.Delete(conn.Id)
			msg := Message{ConnId: 0, Msg: fmt.Sprintf("用户%d离开了房间", conn.Id)}
			v.(*Room).PutMsg(msg)
		}
	}
	conn.mutex.Unlock()
}

//内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
		msg  Message
		room *Room
		ok   bool
		v    interface{}
	)
	for {
		if conn.IsClose {
			goto ERR
		}
		//从websocket中读取数据放入inChan中，如果websocket此时没有数据，则ReadMessage()会阻塞直到有数据过来
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		msg = Message{ConnId: conn.Id, Msg: string(data)}
		//将数据放入房间的msgChan
		v, ok = RoomMap.Load(conn.RoomId)
		if !ok { //如果对应的房间不存在
			goto ERR
		}
		room = v.(*Room)
		room.PutMsg(msg)
		log.Println("房间", conn.RoomId, "：", string(msg.ToBytes()))
	}

ERR:
	fmt.Println("readLoop")
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		msg Message
		err error
	)
	for {
		if conn.IsClose {
			goto ERR
		}
		select {
		//从outChan中读取数据利用wsConn返回给客户端
		case msg = <-conn.outChan:
			if err = conn.wsConn.WriteMessage(websocket.TextMessage, msg.ToBytes()); err != nil {
				goto ERR
			}
		case <-conn.closeChan:
			goto ERR
		}
	}
ERR:
	fmt.Println("writeLoop")
	conn.Close()
}
