package impl

import "fmt"

type Message struct {
	ConnId int
	Msg    string
}

func (msg Message) ToBytes() []byte {
	str := fmt.Sprintf("%d:%v", msg.ConnId, msg.Msg)
	return []byte(str)
}
