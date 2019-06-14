package impl

type Room struct {
	message   chan string   //消息管道
	connMap   []*Connection //用户列表
	userCount int           //当前房间用户数
}
