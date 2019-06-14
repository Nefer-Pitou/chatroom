package impl

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte
	isClose   bool
	mutex     sync.Mutex
	//roomId string //所在房间ID
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}
	//启动读协程
	go conn.readLoop()
	//启动写协程
	go conn.writeLoop()

	return
}

//API
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection was disconnected")
	}
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection was disconnected")
	}
	return
}

func (conn *Connection) Close() {
	//线程安全，可重入的close
	if err := conn.wsConn.Close(); err != nil {
		log.Println(err)
	}
	conn.mutex.Lock()
	if !conn.isClose {
		close(conn.closeChan)
		conn.isClose = true
	}
	conn.mutex.Unlock()
}

//内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		//从websocket中读取数据放入inChan中，如果websocket此时没有数据，则ReadMessage()会阻塞直到有数据过来
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		//fmt.Println("data:", string(data))
		//阻塞在这里，等待inChan有空的位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}

ERR:
	fmt.Println("readLoop")
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		//从outChan中读取数据利用wsConn返回给客户端
		case data = <-conn.outChan:
			if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
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
