package impl

import (
	"log"
	"sync"
)

var (
	//思路：当一个用户连接时，为他分配到一个房间内，
	RoomMap sync.Map
)

type Room struct {
	msgChan chan Message //消息管道
	//connections   []*Connection //用户列表
	connMap   sync.Map
	closeChan chan byte
}

func (r *Room) PutMsg(msg Message) {
	r.msgChan <- msg
}

func InitRoom() (room *Room) {
	room = &Room{
		msgChan:   make(chan Message, 10),
		closeChan: make(chan byte, 1),
	}

	go func() {
		var (
			msg  Message
			err  error
			conn *Connection
		)
		for {
			select {
			case msg = <-room.msgChan:
				//将消息广播到房间内的所有用户
				room.connMap.Range(func(k, v interface{}) bool {
					conn = v.(*Connection)
					if msg.ConnId != k {
						if err = conn.WriteMessage(msg); err != nil {
							log.Println(err)
						}
					}
					return true
				})
			case <-room.closeChan:
				goto ERR
			}

		}
	ERR:
		room.Close()

	}()

	return
}

func (r *Room) Close() error {
	close(r.msgChan)
	close(r.closeChan)
	return nil
}

func (r *Room) PutConn(conn *Connection) {
	r.connMap.Store(conn.Id, conn)
}
